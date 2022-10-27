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
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	faulttolerancePb "github.com/opensergo/opensergo-go/pkg/proto/fault_tolerance/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
)

type MixedRuleCache struct {
	MixedRule
	// map[ruleType] bool, update change status of ruleType
	updateFlagMap map[string]bool
	// map[resourceName] index, flow.Rule index of mixedRuleCache by resourceName, used to update flow.Rule in mixedRuleCache
	flowRuleNameMap map[string]int
}

func newMixedRuleCache() *MixedRuleCache {
	return &MixedRuleCache{
		updateFlagMap: make(map[string]bool),
	}
}

type OpensergoRuleAggregator struct {
	ruleAssembler         RuleAssemblerAggregator
	sentinelUpdateHandler func()

	mixedRuleCache *MixedRuleCache

	// map[kindName] []v1.FaultToleranceRule
	pbTtRuleMapByStrategyKind map[string][]faulttolerancePb.FaultToleranceRule
	// map[kindName] []v1.RateLimitStrategy
	pbRlStrategyMap map[string]faulttolerancePb.RateLimitStrategy
}

func NewOpensergoRuleAggregator() *OpensergoRuleAggregator {
	return &OpensergoRuleAggregator{
		ruleAssembler:  RuleAssemblerAggregator{},
		mixedRuleCache: newMixedRuleCache(),

		pbTtRuleMapByStrategyKind: make(map[string][]faulttolerancePb.FaultToleranceRule),

		pbRlStrategyMap: make(map[string]faulttolerancePb.RateLimitStrategy),
	}
}

func (aggregator *OpensergoRuleAggregator) setSentinelUpdateHandler(sentinelUpdateHandler func()) {
	aggregator.sentinelUpdateHandler = sentinelUpdateHandler
}

var updateFaultToleranceRulesMutex sync.Mutex

func (aggregator *OpensergoRuleAggregator) updateFaultToleranceRules(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	updateFaultToleranceRulesMutex.Lock()
	defer updateFaultToleranceRulesMutex.Unlock()
	for _, pbData := range dataSlice {
		pbFaultToleranceRule := pbData.(*faulttolerancePb.FaultToleranceRule)
		for _, strategyRef := range pbFaultToleranceRule.GetStrategies() {
			kindName := strategyRef.GetKind()
			pbTtRuleSlice := make([]faulttolerancePb.FaultToleranceRule, 0)
			if pbTtRuleSliceLoaded := aggregator.pbTtRuleMapByStrategyKind[kindName]; pbTtRuleSliceLoaded != nil {
				pbTtRuleSlice = pbTtRuleSliceLoaded
			}
			pbTtRuleSlice = append(pbTtRuleSlice, *pbFaultToleranceRule)
			aggregator.pbTtRuleMapByStrategyKind[kindName] = pbTtRuleSlice
		}
	}

	aggregator.updateFlowRule()
	//aggregator.updateCircuitBreakerRule()
	return true, nil
}

var updateRateLimitStrategyMutex sync.Mutex

func (aggregator *OpensergoRuleAggregator) updateRateLimitStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	updateRateLimitStrategyMutex.Lock()
	defer updateRateLimitStrategyMutex.Unlock()
	if len(dataSlice) > 0 {
		for _, pbData := range dataSlice {
			rateLimitStrategy := pbData.(*faulttolerancePb.RateLimitStrategy)
			aggregator.pbRlStrategyMap[rateLimitStrategy.Name] = *rateLimitStrategy
		}
	}

	aggregator.updateFlowRule()
	return true, nil
}

func (aggregator *OpensergoRuleAggregator) updateFlowRule() {
	flowRules := make([]flow.Rule, 0)
	// assembler RateLimitStrategies for FlowRule
	pbRuleOfRateLimitStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefRateLimitStrategy{}.GetSimpleName()]
	flowRulesByRlStrategy := aggregator.ruleAssembler.assembleFlowRulesFromRateLimitStrategies(pbRuleOfRateLimitStrategies, aggregator.pbRlStrategyMap)
	if flowRulesByRlStrategy != nil && len(flowRulesByRlStrategy) > 0 {
		flowRules = append(flowRules, flowRulesByRlStrategy...)
	}
	// TODO assembler other flowRule strategies

	// merge flowRule between cache-data and new-data
	for _, rule := range flowRules {
		flowRuleIndex := aggregator.mixedRuleCache.flowRuleNameMap[rule.ResourceName()]
		// if existed then update
		// else append
		if flowRuleIndex > 0 || (flowRuleIndex == 0 && len(aggregator.mixedRuleCache.FlowRule) == 1) {
			aggregator.mixedRuleCache.FlowRule[flowRuleIndex] = rule
		} else {
			if aggregator.mixedRuleCache.FlowRule == nil {
				aggregator.mixedRuleCache.FlowRule = make([]flow.Rule, 0)
				aggregator.mixedRuleCache.flowRuleNameMap = make(map[string]int)
			}

			aggregator.mixedRuleCache.FlowRule = append(aggregator.mixedRuleCache.FlowRule, rule)
		}
		aggregator.mixedRuleCache.flowRuleNameMap[rule.ResourceName()] = len(aggregator.mixedRuleCache.FlowRule) - 1
	}

	aggregator.mixedRuleCache.updateFlagMap[RuleType_FlowRule] = true
	aggregator.sentinelUpdateHandler()
	aggregator.mixedRuleCache.updateFlagMap[RuleType_FlowRule] = false
}

func (aggregator *OpensergoRuleAggregator) updateCircuitBreakerRule() {
	// TODO add logic of updateCircuitBreakerRule
}
