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

import "github.com/alibaba/sentinel-golang/core/base"

const (
	StatSlotOrder uint32 = 6000
)

var (
	DefaultLLMTokenRatelimitStatSlot = &LLMTokenRatelimitStatSlot{}
)

type LLMTokenRatelimitStatSlot struct {
}

func (s *LLMTokenRatelimitStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (c *LLMTokenRatelimitStatSlot) OnEntryPassed(_ *base.EntryContext) {
	// Do nothing
	return
}

func (c *LLMTokenRatelimitStatSlot) OnEntryBlocked(_ *base.EntryContext, _ *base.BlockError) {
	// Do nothing
	return
}

func (c *LLMTokenRatelimitStatSlot) OnCompleted(ctx *base.EntryContext) {
	if !globalConfig.IsEnabled() {
		return
	}
	usedTokenInfos, ok := ctx.GetPair(KeyUsedTokenInfos).(*UsedTokenInfos)
	if !ok || usedTokenInfos == nil {
		return
	}
	llmTokenRatelimitCtx, ok := ctx.GetPair(KeyContext).(*Context)
	if !ok || llmTokenRatelimitCtx == nil {
		return
	}
	llmTokenRatelimitCtx.Set(KeyUsedTokenInfos, usedTokenInfos)

	rulesInterface := llmTokenRatelimitCtx.Get(KeyMatchedRules)
	if rulesInterface == nil {
		return
	}

	rules, ok := rulesInterface.([]*MatchedRule)
	if !ok || rules == nil {
		return
	}

	for _, rule := range rules {
		globalRuleMatcher.update(llmTokenRatelimitCtx, rule)
	}
}
