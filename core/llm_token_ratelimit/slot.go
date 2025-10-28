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
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	RuleCheckSlotOrder uint32 = 6000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct{}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(resource) == 0 {
		return result
	}

	if passed, rule, snapshot := s.checkPass(ctx); !passed {
		msg := "llm token ratelimit check blocked"
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeLLMTokenRateLimit, msg, rule, snapshot)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeLLMTokenRateLimit, msg, rule, snapshot)
		}
	}

	return result
}

func (s *Slot) checkPass(ctx *base.EntryContext) (bool, *Rule, interface{}) {
	if !globalConfig.IsEnabled() {
		return true, nil, nil
	}
	requestID := generateUUID()
	llmTokenRatelimitCtx, ok := ctx.GetPair(KeyContext).(*Context)
	if !ok || llmTokenRatelimitCtx == nil {
		llmTokenRatelimitCtx = NewContext()
		if llmTokenRatelimitCtx == nil {
			logging.Warn("[LLMTokenRateLimit] failed to create llm token ratelimit context",
				"requestID", requestID,
			)
			return true, nil, nil
		}
		llmTokenRatelimitCtx.extractArgs(ctx)
		llmTokenRatelimitCtx.Set(KeyRequestID, requestID)
		ctx.SetPair(KeyContext, llmTokenRatelimitCtx)
	}
	for _, rule := range getRulesOfResource(ctx.Resource.Name()) {
		if !globalRuleMatcher.checkPass(llmTokenRatelimitCtx, rule) {
			return false, rule, llmTokenRatelimitCtx.Get(KeyResponseHeaders)
		}
	}
	ctx.SetPair(KeyResponseHeaders, llmTokenRatelimitCtx.Get(KeyResponseHeaders))
	return true, nil, nil
}
