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
	"errors"
	"fmt"

	"github.com/alibaba/sentinel-golang/logging"
)

type BaseRuleCollector struct{}

func (c *BaseRuleCollector) Collect(ctx *Context, rule *Rule) []*MatchedRule {
	if c == nil || rule == nil {
		return nil
	}

	reqInfos := extractRequestInfos(ctx) // allow nil for global rate limit

	resourceHash := generateHash(rule.Resource)
	ruleStrategy := rule.Strategy.String()

	estimatedSize := 0
	for _, item := range rule.SpecificItems {
		if item == nil || item.KeyItems == nil {
			continue
		}
		estimatedSize += len(item.KeyItems)
	}

	ruleMap := make(map[string]*MatchedRule, estimatedSize)

	for _, specificItem := range rule.SpecificItems {
		if specificItem == nil || specificItem.KeyItems == nil {
			continue
		}

		identifierChecker := globalRuleMatcher.getIdentifierChecker(specificItem.Identifier.Type)
		if identifierChecker == nil {
			logging.Error(errors.New("unknown identifier.type"),
				"unknown identifier.type in llm_token_ratelimit.BaseRuleCollector.Collect()",
				"identifier.type", specificItem.Identifier.Type.String(),
				"requestID", ctx.Get(KeyRequestID),
			)
			continue
		}

		identifierType := specificItem.Identifier.Type.String()

		for _, keyItem := range specificItem.KeyItems {
			if !identifierChecker.Check(ctx, reqInfos, specificItem.Identifier, keyItem.Key) {
				continue
			}

			timeWindow := keyItem.Time.convertToSeconds()
			if timeWindow == ErrorTimeDuration {
				logging.Error(errors.New("error time window"),
					"error time window in llm_token_ratelimit.BaseRuleCollector.Collect()",
					"requestID", ctx.Get(KeyRequestID),
				)
				continue
			}

			limitKey := fmt.Sprintf(RedisRatelimitKeyFormat,
				resourceHash,
				ruleStrategy,
				identifierType,
				timeWindow,
				keyItem.Token.CountStrategy.String(),
			)
			ruleMap[limitKey] = &MatchedRule{
				Strategy:      rule.Strategy,
				LimitKey:      limitKey,
				TimeWindow:    timeWindow,
				TokenSize:     keyItem.Token.Number,
				CountStrategy: keyItem.Token.CountStrategy,
				// PETA
				Encoding: rule.Encoding,
			}
		}
	}

	rules := make([]*MatchedRule, 0, len(ruleMap))
	for _, rule := range ruleMap {
		rules = append(rules, rule)
	}
	return rules
}
