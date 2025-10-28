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
	"errors"
	"fmt"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

// ================================= FixedWindowUpdater ====================================

//go:embed script/fixed_window/update.lua
var globalFixedWindowUpdateScript string

type FixedWindowUpdater struct{}

func (u *FixedWindowUpdater) Update(ctx *Context, rule *MatchedRule) {
	if u == nil || ctx == nil || rule == nil {
		return
	}

	usedTokenInfos := extractUsedTokenInfos(ctx)
	if usedTokenInfos == nil {
		return
	}

	u.updateLimitKey(ctx, rule, usedTokenInfos)
}

func (u *FixedWindowUpdater) updateLimitKey(ctx *Context, rule *MatchedRule, infos *UsedTokenInfos) {
	if u == nil || ctx == nil || rule == nil || infos == nil {
		return
	}
	calculator := globalTokenCalculator.getCalculator(rule.CountStrategy)
	if calculator == nil {
		logging.Error(errors.New("unknown strategy"),
			"unknown strategy in llm_token_ratelimit.FixedWindowUpdater.updateLimitKey() when get calculator",
			"strategy", rule.CountStrategy.String(),
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
	usedToken := calculator.Calculate(ctx, infos)
	keys := []string{rule.LimitKey}
	args := []interface{}{rule.TokenSize, rule.TimeWindow * 1000, usedToken}
	response, err := globalRedisClient.Eval(globalFixedWindowUpdateScript, keys, args...)
	if err != nil {
		logging.Error(err, "failed to execute redis script in llm_token_ratelimit.FixedWindowUpdater.updateLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
	result := parseRedisResponse(ctx, response)
	if result == nil || len(result) != 2 {
		logging.Error(errors.New("invalid redis response"),
			"invalid redis response in llm_token_ratelimit.FixedWindowUpdater.updateLimitKey()",
			"response", response,
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
}

// ================================= PETAUpdater ====================================

//go:embed script/peta/correct.lua
var globalPETACorrectScript string

type PETAUpdater struct{}

func (u *PETAUpdater) Update(ctx *Context, rule *MatchedRule) {
	if u == nil || ctx == nil || rule == nil {
		return
	}

	usedTokenInfos := extractUsedTokenInfos(ctx)
	if usedTokenInfos == nil {
		return
	}

	u.updateLimitKey(ctx, rule, usedTokenInfos)
}

func (u *PETAUpdater) updateLimitKey(ctx *Context, rule *MatchedRule, infos *UsedTokenInfos) {
	if u == nil || ctx == nil || rule == nil || infos == nil {
		return
	}

	calculator := globalTokenCalculator.getCalculator(rule.CountStrategy)
	if calculator == nil {
		logging.Error(errors.New("unknown strategy"),
			"unknown strategy in llm_token_ratelimit.PETAUpdater.updateLimitKey() when get calculator",
			"strategy", rule.CountStrategy.String(),
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
	actualToken := calculator.Calculate(ctx, infos)

	slidingWindowKey := fmt.Sprintf(PETASlidingWindowKeyFormat, generateHash(rule.LimitKey), rule.LimitKey)
	tokenBucketKey := fmt.Sprintf(PETATokenBucketKeyFormat, generateHash(rule.LimitKey), rule.LimitKey)
	tokenEncoderKey := fmt.Sprintf(TokenEncoderKeyFormat, generateHash(rule.LimitKey), rule.Encoding.Provider.String(), rule.Encoding.Model, rule.LimitKey)

	keys := []string{slidingWindowKey, tokenBucketKey, tokenEncoderKey}
	args := []interface{}{rule.EstimatedToken, util.CurrentTimeMillis(), rule.TokenSize, rule.TimeWindow * 1000, actualToken, generateRandomString(PETARandomStringLength)}
	response, err := globalRedisClient.Eval(globalPETACorrectScript, keys, args...)
	if err != nil {
		logging.Error(err, "failed to execute redis script in llm_token_ratelimit.PETAUpdater.updateLimitKey()",
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
	result := parseRedisResponse(ctx, response)
	if result == nil || len(result) != 1 {
		logging.Error(errors.New("invalid redis response"),
			"invalid redis response in llm_token_ratelimit.PETAUpdater.updateLimitKey()",
			"response", response,
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}

	correctResult := result[0]
	if correctResult != PETACorrectOK && correctResult != PETACorrectOverestimateError { // Temporarily unable to handle overestimation cases
		logging.Warn("[LLMTokenRateLimit] failed to update the limit key",
			"correct_result", correctResult,
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}

	RecordMetric(MetricItem{
		Timestamp:      util.CurrentTimeMillis(),
		RequestID:      ctx.Get(KeyRequestID).(string),
		EstimatedToken: rule.EstimatedToken,
		ActualToken:    actualToken,
		CorrectResult:  result[0],
	})
}
