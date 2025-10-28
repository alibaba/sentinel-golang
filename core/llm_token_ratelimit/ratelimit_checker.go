// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package llm_token_ratelimit

import (
	_ "embed"
	"fmt"
	"time"

	"errors"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

// ================================= FixedWindowChecker ====================================

//go:embed script/fixed_window/query.lua
var globalFixedWindowQueryScript string

type FixedWindowChecker struct{}

func (c *FixedWindowChecker) Check(ctx *Context, rules []*MatchedRule) bool {
	if c == nil {
		return true
	}

	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		if !c.checkLimitKey(ctx, rule) {
			return false
		}
	}
	return true
}

func (c *FixedWindowChecker) checkLimitKey(ctx *Context, rule *MatchedRule) bool {
	if c == nil {
		return true
	}

	keys := []string{rule.LimitKey}
	args := []interface{}{rule.TokenSize, rule.TimeWindow * 1000}
	response, err := globalRedisClient.Eval(globalFixedWindowQueryScript, keys, args...)
	if err != nil {
		logging.Error(err, "failed to execute redis script in llm_token_ratelimit.FixedWindowChecker.checkLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}
	result := parseRedisResponse(ctx, response)
	if result == nil || len(result) != 2 {
		logging.Error(errors.New("invalid redis response"),
			"invalid redis response in llm_token_ratelimit.FixedWindowChecker.checkLimitKey()",
			"response", response,
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}

	remaining := result[0]
	responseHeader := NewResponseHeader()
	if responseHeader == nil {
		logging.Error(errors.New("failed to create response header"),
			"failed to create response header in llm_token_ratelimit.FixedWindowChecker.checkLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}
	defer func() {
		ctx.Set(KeyResponseHeaders, responseHeader)
	}()
	// set response headers
	responseHeader.Set(ResponseHeaderRequestID, ctx.Get(KeyRequestID).(string))
	responseHeader.Set(ResponseHeaderRemainingTokens, fmt.Sprintf("%d", remaining))
	if remaining < 0 {
		// set waiting time in milliseconds
		responseHeader.Set(ResponseHeaderWaitingTime, (time.Duration(result[1]) * time.Millisecond).String())
		// set error code and message
		responseHeader.ErrorCode = globalConfig.GetErrorCode()
		responseHeader.ErrorMessage = globalConfig.GetErrorMsg()
		// reject the request
		return false
	}
	return true
}

// ================================= PETAChecker ====================================

//go:embed script/peta/withhold.lua
var globalPETAWithholdScript string

type PETAChecker struct{}

func (c *PETAChecker) Check(ctx *Context, rules []*MatchedRule) bool {
	if c == nil || ctx == nil {
		return true
	}

	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		if !c.checkLimitKey(ctx, rule) {
			return false
		}
	}
	return true
}

func (c *PETAChecker) checkLimitKey(ctx *Context, rule *MatchedRule) bool {
	if c == nil || ctx == nil || rule == nil {
		return true
	}

	prompts := []string{}
	reqInfos := extractRequestInfos(ctx)
	if reqInfos != nil {
		prompts = reqInfos.Prompts
	}

	length, err := c.countTokens(ctx, prompts, rule)
	if err != nil {
		logging.Error(err, "failed to count tokens in llm_token_ratelimit.PETAChecker.checkLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}

	slidingWindowKey := fmt.Sprintf(PETASlidingWindowKeyFormat, generateHash(rule.LimitKey), rule.LimitKey)
	tokenBucketKey := fmt.Sprintf(PETATokenBucketKeyFormat, generateHash(rule.LimitKey), rule.LimitKey)
	tokenEncoderKey := fmt.Sprintf(TokenEncoderKeyFormat, generateHash(rule.LimitKey), rule.Encoding.Provider.String(), rule.Encoding.Model, rule.LimitKey)

	keys := []string{slidingWindowKey, tokenBucketKey, tokenEncoderKey}
	args := []interface{}{length, util.CurrentTimeMillis(), rule.TokenSize, rule.TimeWindow * 1000, generateRandomString(PETARandomStringLength)}
	response, err := globalRedisClient.Eval(globalPETAWithholdScript, keys, args...)
	if err != nil {
		logging.Error(err, "failed to execute redis script in llm_token_ratelimit.PETAChecker.checkLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}
	result := parseRedisResponse(ctx, response)
	if result == nil || len(result) != 4 {
		logging.Error(errors.New("invalid redis response"),
			"invalid redis response in llm_token_ratelimit.PETAChecker.checkLimitKey()",
			"response", response,
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}

	RecordMetric(MetricItem{
		Timestamp:          util.CurrentTimeMillis(),
		RequestID:          ctx.Get(KeyRequestID).(string),
		LimitKey:           rule.LimitKey,
		CurrentCapacity:    result[0],
		WaitingTime:        result[1],
		EstimatedToken:     result[2],
		Difference:         result[3],
		TokenizationLength: length,
	})

	// TODO: add waiting and timeout callback
	waitingTime := result[1]
	responseHeader := NewResponseHeader()
	if responseHeader == nil {
		logging.Error(errors.New("failed to create response header"),
			"failed to create response header in llm_token_ratelimit.PETAChecker.checkLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}
	defer func() {
		ctx.Set(KeyResponseHeaders, responseHeader)
	}()
	// set response headers
	responseHeader.Set(ResponseHeaderRequestID, ctx.Get(KeyRequestID).(string))
	responseHeader.Set(ResponseHeaderRemainingTokens, fmt.Sprintf("%d", result[0]))
	if waitingTime != PETANoWaiting {
		// set waiting time in milliseconds
		responseHeader.Set(ResponseHeaderWaitingTime, (time.Duration(waitingTime) * time.Millisecond).String())
		// set error code and message
		responseHeader.ErrorCode = globalConfig.GetErrorCode()
		responseHeader.ErrorMessage = globalConfig.GetErrorMsg()
		// reject the request
		return false
	}
	c.cacheEstimatedToken(rule, result[2])
	return true
}

func (c *PETAChecker) countTokens(ctx *Context, prompts []string, rule *MatchedRule) (int, error) {
	if c == nil {
		return 0, fmt.Errorf("PETAChecker is nil")
	}

	switch rule.CountStrategy {
	case OutputTokens: // cannot predict output tokens
		return 0, nil
	case InputTokens, TotalTokens:
		encoder := LookupTokenEncoder(ctx, rule.Encoding) // try to get cached encoder
		if encoder == nil {
			encoder = NewTokenEncoder(ctx, rule.Encoding) // create a new encoder
			if encoder == nil {
				return 0, fmt.Errorf("failed to create token encoder for encoding")
			}
		}
		length, err := encoder.CountTokens(ctx, prompts, rule)
		if err != nil {
			return 0, fmt.Errorf("failed to count tokens: %v", err)
		}
		return length, nil
	}
	return 0, fmt.Errorf("unknown count strategy: %s", rule.CountStrategy.String())
}

func (c *PETAChecker) cacheEstimatedToken(rule *MatchedRule, count int64) {
	if c == nil || rule == nil {
		return
	}
	rule.EstimatedToken = count
}
