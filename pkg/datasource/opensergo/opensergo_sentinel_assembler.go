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
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	faulttolerancePb "github.com/opensergo/opensergo-go/pkg/proto/fault_tolerance/v1"
	"github.com/pkg/errors"
)

type OpensergoSentinelRuleAssembler struct {
}

func (assembler OpensergoSentinelRuleAssembler) assembleFlowRulesFromRateLimitStrategies(pbRules []*faulttolerancePb.FaultToleranceRule, pbStrategyMap map[string]*faulttolerancePb.RateLimitStrategy) []flow.Rule {
	if len(pbRules) == 0 {
		return []flow.Rule{}
	}
	var flowRules []flow.Rule
	for _, rule := range pbRules {
		var strategies []*faulttolerancePb.RateLimitStrategy
		for _, strategy := range rule.GetStrategies() {
			limitStrategy := configkind.ConfigKindRefRateLimitStrategy{}.GetSimpleName()
			if strategy.Kind == limitStrategy {
				strategies = append(strategies, pbStrategyMap[strategy.Name])
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
				flowRule = assembler.fillFlowRuleWithRateLimitStrategy(flowRule, strategy)
				if flowRule != nil {
					flowRules = append(flowRules, *flowRule)
				}
			}
		}
	}
	return flowRules
}

func (assembler OpensergoSentinelRuleAssembler) fillFlowRuleWithRateLimitStrategy(flowRule *flow.Rule, pbStrategy *faulttolerancePb.RateLimitStrategy) *flow.Rule {
	if flowRule == nil || pbStrategy == nil {
		return flowRule
	}

	defer func() *flow.Rule {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, "[OpenSergoDatasource] Ignoring OpenSergo RateLimitStrategy due to covert failure.", "resourceName", flowRule.Resource, "pbStrategy", pbStrategy)
		}
		return nil
	}()

	// TODO fill field-mapping between sentinel-rule and pb-message
	flowRule.Threshold = float64(pbStrategy.Threshold)
	flowRule.TokenCalculateStrategy = flow.Direct
	flowRule.ControlBehavior = flow.Reject

	return flowRule
}

func (assembler OpensergoSentinelRuleAssembler) assembleFlowRulesFromThrottlingStrategies(pbRules []*faulttolerancePb.FaultToleranceRule, pbStrategyMap map[string]*faulttolerancePb.ThrottlingStrategy) []flow.Rule {
	if len(pbRules) == 0 {
		return []flow.Rule{}
	}
	var flowRules []flow.Rule
	for _, rule := range pbRules {
		var strategies []*faulttolerancePb.ThrottlingStrategy
		for _, strategy := range rule.GetStrategies() {
			throttlingStrategy := configkind.ConfigKindRefThrottlingStrategy{}.GetSimpleName()
			if strategy.Kind == throttlingStrategy {
				strategies = append(strategies, pbStrategyMap[strategy.Name])
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
				flowRule = assembler.fillFlowRuleWithThrottlingStrategy(flowRule, strategy)
				if flowRule != nil {
					flowRules = append(flowRules, *flowRule)
				}
			}
		}
	}
	return flowRules
}

func (assembler OpensergoSentinelRuleAssembler) fillFlowRuleWithThrottlingStrategy(flowRule *flow.Rule, pbStrategy *faulttolerancePb.ThrottlingStrategy) *flow.Rule {
	if flowRule == nil || pbStrategy == nil {
		return flowRule
	}

	defer func() *flow.Rule {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, "[OpenSergoDatasource] Ignoring OpenSergo ThrottlingStrategy due to covert failure.", "resourceName", flowRule.Resource, "pbStrategy", pbStrategy)
		}
		return nil
	}()

	// TODO fill field-mapping between sentinel-rule and pb-message
	flowRule.Threshold = float64(1000 / pbStrategy.MinIntervalMillisOfRequests)
	flowRule.TokenCalculateStrategy = flow.Direct
	flowRule.ControlBehavior = flow.Throttling
	flowRule.MaxQueueingTimeMs = uint32(pbStrategy.QueueTimeoutMillis)

	return flowRule
}

func (assembler OpensergoSentinelRuleAssembler) assembleIsolationRulesFromConcurrencyLimitStrategies(pbRules []*faulttolerancePb.FaultToleranceRule, pbStrategyMap map[string]*faulttolerancePb.ConcurrencyLimitStrategy) []isolation.Rule {
	if len(pbRules) == 0 {
		return []isolation.Rule{}
	}
	var isolationRules []isolation.Rule
	for _, rule := range pbRules {
		var strategies []*faulttolerancePb.ConcurrencyLimitStrategy
		for _, strategy := range rule.GetStrategies() {
			concurrencyLimitStrategy := configkind.ConfigKindRefConcurrencyLimitStrategy{}.GetSimpleName()
			if strategy.Kind == concurrencyLimitStrategy {
				strategies = append(strategies, pbStrategyMap[strategy.Name])
			}
		}
		if len(strategies) == 0 {
			continue
		}

		for _, targetRef := range rule.GetTargets() {
			resourceName := targetRef.TargetResourceName

			for _, strategy := range strategies {
				isolationRule := new(isolation.Rule)
				isolationRule.Resource = resourceName
				isolationRule = assembler.fillIsolationRuleWithConcurrencyLimitStrategy(isolationRule, strategy)
				if isolationRule != nil {
					isolationRules = append(isolationRules, *isolationRule)
				}
			}
		}
	}
	return isolationRules
}

func (assembler OpensergoSentinelRuleAssembler) fillIsolationRuleWithConcurrencyLimitStrategy(isolationRule *isolation.Rule, pbStrategy *faulttolerancePb.ConcurrencyLimitStrategy) *isolation.Rule {
	if isolationRule == nil || pbStrategy == nil {
		return isolationRule
	}

	defer func() *flow.Rule {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, "[OpenSergoDatasource] Ignoring OpenSergo ConcurrencyLimitStrategy due to covert failure.", "resourceName", isolationRule.Resource, "pbStrategy", pbStrategy)
		}
		return nil
	}()

	// TODO fill field-mapping between sentinel-rule and pb-message
	isolationRule.Threshold = uint32(pbStrategy.MaxConcurrency)
	isolationRule.MetricType = isolation.Concurrency

	return isolationRule
}

func (assembler OpensergoSentinelRuleAssembler) assembleCircuitBreakerRulesFromCircuitBreakerStrategies(pbRules []*faulttolerancePb.FaultToleranceRule, pbStrategyMap map[string]*faulttolerancePb.CircuitBreakerStrategy) []circuitbreaker.Rule {
	if len(pbRules) == 0 {
		return []circuitbreaker.Rule{}
	}
	var circuitbreakerRules []circuitbreaker.Rule
	for _, rule := range pbRules {
		var strategies []*faulttolerancePb.CircuitBreakerStrategy
		for _, strategy := range rule.GetStrategies() {
			circuitBreakerStrategy := configkind.ConfigKindRefCircuitBreakerStrategy{}.GetSimpleName()
			if strategy.Kind == circuitBreakerStrategy {
				strategies = append(strategies, pbStrategyMap[strategy.Name])
			}
		}
		if len(strategies) == 0 {
			continue
		}

		for _, targetRef := range rule.GetTargets() {
			resourceName := targetRef.TargetResourceName

			for _, strategy := range strategies {
				circuitbreakerRule := new(circuitbreaker.Rule)
				circuitbreakerRule.Resource = resourceName
				circuitbreakerRule = assembler.fillCircuitbreakerRuleWithCircuitBreakerStrategy(circuitbreakerRule, strategy)
				if circuitbreakerRule != nil {
					circuitbreakerRules = append(circuitbreakerRules, *circuitbreakerRule)
				}
			}
		}
	}
	return circuitbreakerRules
}

func (assembler OpensergoSentinelRuleAssembler) fillCircuitbreakerRuleWithCircuitBreakerStrategy(circuitbreakerRule *circuitbreaker.Rule, pbStrategy *faulttolerancePb.CircuitBreakerStrategy) *circuitbreaker.Rule {
	if circuitbreakerRule == nil || pbStrategy == nil {
		return circuitbreakerRule
	}

	defer func() *circuitbreaker.Rule {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, "[OpenSergoDatasource] Ignoring OpenSergo CircuitBreakerStrategy due to covert failure.", "resourceName", circuitbreakerRule.Resource, "pbStrategy", pbStrategy)
		}
		return nil
	}()

	// TODO fill field-mapping between sentinel-rule and pb-message
	switch pbStrategy.Strategy {
	case faulttolerancePb.CircuitBreakerStrategy_STRATEGY_SLOW_REQUEST_RATIO:
		circuitbreakerRule.Threshold = float64(pbStrategy.SlowCondition.MaxAllowedRtMillis)
		break
	case faulttolerancePb.CircuitBreakerStrategy_STRATEGY_ERROR_REQUEST_RATIO:
		circuitbreakerRule.Threshold = pbStrategy.TriggerRatio
		break
	default:
		logging.Info("[OpenSergoDatasource] unknow CircuitBreakerStrategy.", "resourceName", circuitbreakerRule.Resource, "pbStrategy", pbStrategy)
	}
	circuitbreakerRule.MinRequestAmount = uint64(pbStrategy.MinRequestAmount)

	return circuitbreakerRule
}
