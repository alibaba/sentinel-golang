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
	"sync"

	"github.com/alibaba/sentinel-golang/core/isolation"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	faulttolerancePb "github.com/opensergo/opensergo-go/pkg/proto/fault_tolerance/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MixedRuleCache struct {
	MixedRule
	// map[ruleType] bool, update change status of ruleType
	updateFlagMap map[string]bool
}

func newMixedRuleCache() *MixedRuleCache {
	return &MixedRuleCache{
		updateFlagMap: make(map[string]bool),
	}
}

type OpensergoRuleAggregator struct {
	ruleAssembler OpensergoSentinelRuleAssembler

	sentinelUpdateMutex   sync.Mutex
	sentinelUpdateHandler func() error

	mixedRuleCache *MixedRuleCache

	// map[kindName] []v1.FaultToleranceRule
	// store and update FaultToleranceRule from protobufMessage by kindName
	pbTtRuleMapByStrategyKind map[string][]*faulttolerancePb.FaultToleranceRule
	// map[kindName] []v1.RateLimitStrategy
	// store and update RateLimitStrategy from protobufMessage by kindName
	pbRlStrategyMap map[string]*faulttolerancePb.RateLimitStrategy
	// map[kindName] []v1.ThrottlingStrategy
	// store and update ThrottlingStrategy from protobufMessage by kindName
	pbThlStrategyMap map[string]*faulttolerancePb.ThrottlingStrategy
	// map[kindName] []v1.ConcurrencyLimitStrategy
	// store and update ConcurrencyLimitStrategy from protobufMessage by kindName
	pbClStrategyMap map[string]*faulttolerancePb.ConcurrencyLimitStrategy
	// map[kindName] []v1.CircuitBreakerStrategy
	// store and update CircuitBreakerStrategy from protobufMessage by kindName
	pbCbStrategyMap map[string]*faulttolerancePb.CircuitBreakerStrategy
}

func NewOpensergoRuleAggregator() *OpensergoRuleAggregator {
	return &OpensergoRuleAggregator{
		ruleAssembler:  OpensergoSentinelRuleAssembler{},
		mixedRuleCache: newMixedRuleCache(),

		pbTtRuleMapByStrategyKind: make(map[string][]*faulttolerancePb.FaultToleranceRule),
		pbRlStrategyMap:           make(map[string]*faulttolerancePb.RateLimitStrategy),
	}
}

func (aggregator *OpensergoRuleAggregator) setSentinelUpdateHandler(sentinelUpdateHandler func() error) {
	aggregator.sentinelUpdateHandler = sentinelUpdateHandler
}

// doSentinelUpdateHandler update into sentinel with sync.Mutex.
func (aggregator *OpensergoRuleAggregator) doSentinelUpdateHandler(ruleType string) {
	aggregator.sentinelUpdateMutex.Lock()
	defer aggregator.sentinelUpdateMutex.Unlock()

	aggregator.mixedRuleCache.updateFlagMap[ruleType] = true
	if err := aggregator.sentinelUpdateHandler(); err != nil {
		// TODO handle error
		return
	}
	aggregator.mixedRuleCache.updateFlagMap[ruleType] = false
}

// updateFaultToleranceRules store and update FaultToleranceRules from protobufMessage
func (aggregator *OpensergoRuleAggregator) updateFaultToleranceRules(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	for _, pbData := range dataSlice {
		pbFaultToleranceRule := pbData.(*faulttolerancePb.FaultToleranceRule)
		for _, strategyRef := range pbFaultToleranceRule.GetStrategies() {
			kindName := strategyRef.GetKind()
			pbTtRuleSlice := make([]*faulttolerancePb.FaultToleranceRule, 0)
			if pbTtRuleSliceLoaded := aggregator.pbTtRuleMapByStrategyKind[kindName]; pbTtRuleSliceLoaded != nil {
				pbTtRuleSlice = pbTtRuleSliceLoaded
			}
			pbTtRuleSlice = append(pbTtRuleSlice, pbFaultToleranceRule)
			aggregator.pbTtRuleMapByStrategyKind[kindName] = pbTtRuleSlice
		}
	}

	aggregator.handleFlowRuleUpdate()
	aggregator.handleCircuitBreakerRuleUpdate()
	return true, nil
}

// updateRateLimitStrategy store and update RateLimitStrategy from protobufMessage
func (aggregator *OpensergoRuleAggregator) updateRateLimitStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	if len(dataSlice) > 0 {
		aggregator.pbRlStrategyMap = make(map[string]*faulttolerancePb.RateLimitStrategy)
		for _, pbData := range dataSlice {
			rateLimitStrategy := pbData.(*faulttolerancePb.RateLimitStrategy)
			aggregator.pbRlStrategyMap[rateLimitStrategy.Name] = rateLimitStrategy
		}
	}

	aggregator.handleFlowRuleUpdate()
	return true, nil
}

// updateThrottlingStrategy store and update ThrottlingStrategy from protobufMessage
func (aggregator *OpensergoRuleAggregator) updateThrottlingStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	if len(dataSlice) > 0 {
		aggregator.pbThlStrategyMap = make(map[string]*faulttolerancePb.ThrottlingStrategy)
		for _, pbData := range dataSlice {
			throttlingStrategy := pbData.(*faulttolerancePb.ThrottlingStrategy)
			aggregator.pbThlStrategyMap[throttlingStrategy.Name] = throttlingStrategy
		}
	}

	aggregator.handleFlowRuleUpdate()
	return true, nil
}

// updateConcurrencyLimitStrategy store and update ConcurrencyLimitStrategy from protobufMessage
func (aggregator *OpensergoRuleAggregator) updateConcurrencyLimitStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	if len(dataSlice) > 0 {
		aggregator.pbClStrategyMap = make(map[string]*faulttolerancePb.ConcurrencyLimitStrategy)
		for _, pbData := range dataSlice {
			concurrencyLimitStrategy := pbData.(*faulttolerancePb.ConcurrencyLimitStrategy)
			aggregator.pbClStrategyMap[concurrencyLimitStrategy.Name] = concurrencyLimitStrategy
		}
	}

	aggregator.handleIsolationRuleUpdate()
	return true, nil
}

// handleFlowRuleUpdate assemble into FlowRule for Sentinel, and load into Sentinel.
func (aggregator *OpensergoRuleAggregator) handleFlowRuleUpdate() {
	flowRules := make([]flow.Rule, 0)
	// assembler RateLimitStrategies for FlowRule
	pbRuleOfRateLimitStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefRateLimitStrategy{}.GetSimpleName()]
	flowRulesByRlStrategy := aggregator.ruleAssembler.assembleFlowRulesFromRateLimitStrategies(pbRuleOfRateLimitStrategies, aggregator.pbRlStrategyMap)
	if flowRulesByRlStrategy != nil && len(flowRulesByRlStrategy) > 0 {
		flowRules = append(flowRules, flowRulesByRlStrategy...)
	}
	// assembler ThrottlingStrategy for flowRule
	pbRuleOfThrottlingStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefThrottlingStrategy{}.GetSimpleName()]
	flowRulesByThlStrategy := aggregator.ruleAssembler.assembleFlowRulesFromThrottlingStrategies(pbRuleOfThrottlingStrategies, aggregator.pbThlStrategyMap)
	if flowRulesByThlStrategy != nil && len(flowRulesByThlStrategy) > 0 {
		flowRules = append(flowRules, flowRulesByThlStrategy...)
	}

	// reset flowRule
	aggregator.mixedRuleCache.FlowRule = flowRules
	// do SentinelUpdate with mutex lock.
	aggregator.doSentinelUpdateHandler(RuleType_FlowRule)
}

// handleIsolationRuleUpdate assemble into IsolationRule for Sentinel, and load into Sentinel.
func (aggregator *OpensergoRuleAggregator) handleIsolationRuleUpdate() {
	isolationRules := make([]isolation.Rule, 0)
	// assembler ConcurrencyLimitStrategy for IsolationRule
	pbRuleOfConcurrencyLimitStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefConcurrencyLimitStrategy{}.GetSimpleName()]
	isolationRulesRulesByClStrategy := aggregator.ruleAssembler.assembleIsolationRulesFromConcurrencyLimitStrategies(pbRuleOfConcurrencyLimitStrategies, aggregator.pbClStrategyMap)
	if isolationRulesRulesByClStrategy != nil && len(isolationRulesRulesByClStrategy) > 0 {
		isolationRules = append(isolationRules, isolationRulesRulesByClStrategy...)
	}

	// reset flowRule
	aggregator.mixedRuleCache.IsolationRule = isolationRules
	// do SentinelUpdate with mutex lock.
	aggregator.doSentinelUpdateHandler(RuleType_IsolationRule)
}

// updateCircuitBreakerStrategy store and update CircuitBreakerStrategy from protobufMessage
func (aggregator *OpensergoRuleAggregator) updateCircuitBreakerStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	if len(dataSlice) > 0 {
		aggregator.pbCbStrategyMap = make(map[string]*faulttolerancePb.CircuitBreakerStrategy)
		for _, pbData := range dataSlice {
			circuitBreakerStrategy := pbData.(*faulttolerancePb.CircuitBreakerStrategy)
			aggregator.pbCbStrategyMap[circuitBreakerStrategy.Name] = circuitBreakerStrategy
		}
	}

	aggregator.handleCircuitBreakerRuleUpdate()
	return true, nil
}

// handleCircuitBreakerRuleUpdate assemble into CircuitBreakerRule for Sentinel, and load into Sentinel.
func (aggregator *OpensergoRuleAggregator) handleCircuitBreakerRuleUpdate() {
	circuitBreakerRules := make([]circuitbreaker.Rule, 0)
	// assembler CircuitBreakerStrategy for flowRule
	pbRuleOfCircuitBreakerStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefCircuitBreakerStrategy{}.GetSimpleName()]
	circuitBreakerRuleByCbStrategy := aggregator.ruleAssembler.assembleCircuitBreakerRulesFromCircuitBreakerStrategies(pbRuleOfCircuitBreakerStrategies, aggregator.pbCbStrategyMap)
	if circuitBreakerRuleByCbStrategy != nil && len(circuitBreakerRuleByCbStrategy) > 0 {
		circuitBreakerRules = append(circuitBreakerRules, circuitBreakerRuleByCbStrategy...)
	}

	// reset flowRule
	aggregator.mixedRuleCache.CircuitBreakerRule = circuitBreakerRules
	// do SentinelUpdate with mutex lock.
	aggregator.doSentinelUpdateHandler(RuleType_CircuitBreakerRule)

}
