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

	"github.com/alibaba/sentinel-golang/logging"
)

type MatchedRule struct {
	Strategy      Strategy
	LimitKey      string
	TimeWindow    int64 // seconds
	TokenSize     int64
	CountStrategy CountStrategy
	// PETA
	Encoding       TokenEncoding
	EstimatedToken int64
}

type MatchedRuleCollector interface {
	Collect(ctx *Context, rule *Rule) []*MatchedRule
}

type RateLimitChecker interface {
	Check(ctx *Context, rules []*MatchedRule) bool
}

type IdentifierChecker interface {
	Check(ctx *Context, infos *RequestInfos, identifier Identifier, pattern string) bool
}

type TokenUpdater interface {
	Update(ctx *Context, rule *MatchedRule)
}

var globalRuleMatcher = NewDefaultRuleMatcher()

type RuleMatcher struct {
	MatchedRuleCollectors map[Strategy]MatchedRuleCollector
	RateLimitCheckers     map[Strategy]RateLimitChecker
	IdentifierCheckers    map[IdentifierType]IdentifierChecker
	TokenUpdaters         map[Strategy]TokenUpdater
}

func NewDefaultRuleMatcher() *RuleMatcher {
	return &RuleMatcher{
		MatchedRuleCollectors: map[Strategy]MatchedRuleCollector{
			FixedWindow: &BaseRuleCollector{},
			PETA:        &BaseRuleCollector{},
		},
		RateLimitCheckers: map[Strategy]RateLimitChecker{
			FixedWindow: &FixedWindowChecker{},
			PETA:        &PETAChecker{},
		},
		TokenUpdaters: map[Strategy]TokenUpdater{
			FixedWindow: &FixedWindowUpdater{},
			PETA:        &PETAUpdater{},
		},
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{
			AllIdentifier: &AllIdentifierChecker{},
			Header:        &HeaderChecker{},
		},
	}
}

func (m *RuleMatcher) getMatchedRuleCollector(strategy Strategy) MatchedRuleCollector {
	if m == nil {
		return nil
	}
	collector, exists := m.MatchedRuleCollectors[strategy]
	if !exists {
		return nil
	}
	return collector
}

func (m *RuleMatcher) getRateLimitChecker(strategy Strategy) RateLimitChecker {
	if m == nil {
		return nil
	}
	checker, exists := m.RateLimitCheckers[strategy]
	if !exists {
		return nil
	}
	return checker
}

func (m *RuleMatcher) getTokenUpdater(strategy Strategy) TokenUpdater {
	if m == nil {
		return nil
	}
	updater, exists := m.TokenUpdaters[strategy]
	if !exists {
		return nil
	}
	return updater
}

func (m *RuleMatcher) getIdentifierChecker(identifier IdentifierType) IdentifierChecker {
	if m == nil {
		return nil
	}
	checker, exists := m.IdentifierCheckers[identifier]
	if !exists {
		return nil
	}
	return checker
}

func (m *RuleMatcher) checkPass(ctx *Context, rule *Rule) bool {
	if m == nil {
		return true
	}
	collector := m.getMatchedRuleCollector(rule.Strategy)
	if collector == nil {
		logging.Error(errors.New("unknown strategy"),
			"unknown strategy in llm_token_ratelimit.RuleMatcher.checkPass() when get collector",
			"strategy", rule.Strategy.String(),
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}

	rules := collector.Collect(ctx, rule)
	if len(rules) == 0 {
		return true
	}

	checker := m.getRateLimitChecker(rule.Strategy)
	if checker == nil {
		logging.Error(errors.New("unknown strategy"),
			"unknown strategy in llm_token_ratelimit.RuleMatcher.checkPass() when get checker",
			"strategy", rule.Strategy.String(),
			"requestID", ctx.Get(KeyRequestID),
		)
		return true
	}

	if passed := checker.Check(ctx, rules); !passed {
		return false
	}

	m.cacheMatchedRules(ctx, rules)
	return true
}

func (m *RuleMatcher) cacheMatchedRules(ctx *Context, newRules []*MatchedRule) {
	if m == nil {
		return
	}

	if len(newRules) == 0 {
		return
	}

	existingValue := ctx.Get(KeyMatchedRules)
	if existingValue == nil {
		ctx.Set(KeyMatchedRules, newRules)
	} else {
		existingRules, ok := existingValue.([]*MatchedRule)
		if !ok || existingRules == nil {
			ctx.Set(KeyMatchedRules, newRules)
		} else {
			allRules := make([]*MatchedRule, 0, len(existingRules)+len(newRules))
			allRules = append(allRules, existingRules...)
			allRules = append(allRules, newRules...)
			ctx.Set(KeyMatchedRules, allRules)
		}
	}
}

func (m *RuleMatcher) update(ctx *Context, rule *MatchedRule) {
	if m == nil {
		return
	}

	if rule == nil {
		return
	}
	updater := m.getTokenUpdater(rule.Strategy)
	if updater == nil {
		logging.Error(errors.New("unknown strategy"),
			"unknown strategy in llm_token_ratelimit.RuleMatcher.update() when get updater",
			"strategy", rule.Strategy.String(),
			"requestID", ctx.Get(KeyRequestID),
		)
		return
	}
	updater.Update(ctx, rule)
}
