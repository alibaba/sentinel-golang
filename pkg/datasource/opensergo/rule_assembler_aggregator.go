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

package opensergo

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	faulttolerancePb "github.com/opensergo/opensergo-go/pkg/proto/fault_tolerance/v1"
	"github.com/pkg/errors"
)

type RuleAssemblerAggregator struct {
}

func (assembler RuleAssemblerAggregator) assembleFlowRulesFromRateLimitStrategies(pbRules []faulttolerancePb.FaultToleranceRule, pbRlStrategyMap map[string]faulttolerancePb.RateLimitStrategy) []flow.Rule {
	if len(pbRules) == 0 {
		return []flow.Rule{}
	}
	var flowRules []flow.Rule
	for _, rule := range pbRules {
		var strategies []faulttolerancePb.RateLimitStrategy
		for _, strategy := range rule.GetStrategies() {
			limitStrategy := configkind.ConfigKindRefRateLimitStrategy{}.GetSimpleName()
			if strategy.Kind == limitStrategy {
				strategies = append(strategies, pbRlStrategyMap[strategy.Name])
			}
		}
		if len(strategies) == 0 {
			continue
		}

		for _, targetRef := range rule.GetTargets() {
			resourceName := targetRef.TargetResourceName

			for _, strategy := range strategies {
				flowRule := new(flow.Rule)
				flowRule.Resource = resourceName
				flowRule = fillFlowRuleWithRateLimitStrategy(flowRule, strategy)
				if flowRule != nil {
					flowRules = append(flowRules, *flowRule)
				}
			}
		}
	}
	return flowRules
}

func fillFlowRuleWithRateLimitStrategy(flowRule *flow.Rule, strategy faulttolerancePb.RateLimitStrategy) *flow.Rule {
	if flowRule == nil {
		return flowRule
	}

	defer func() *flow.Rule {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, fmt.Sprintf("Ignoring OpenSergo RateLimitStrategy due to covert failure, resourceName=%v, strategy=%v", flowRule.Resource, strategy))
		}
		return nil
	}()

	// TODO fill field-mapping between sentinel-rule and pb-message
	flowRule.Threshold = float64(strategy.Threshold)
	flowRule.TokenCalculateStrategy = flow.Direct
	flowRule.ControlBehavior = flow.Reject

	return flowRule
}
