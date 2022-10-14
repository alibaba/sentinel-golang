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

type OpensergoRuleAggregator struct {
	// map[kindName] []flow.Rule
	dataMap map[string][]flow.Rule

	// map[kindName] []v1.FaultToleranceRule
	pbTtRuleMapByStrategyKind map[string][]faulttolerancePb.FaultToleranceRule
	// map[kindName] []v1.RateLimitStrategy
	pbRlStrategyMap map[string]faulttolerancePb.RateLimitStrategy

	ruleAssembler         RuleAssemblerAggregator
	sentinelUpdateHandler func()
}

func NewOpensergoRuleAggregator() *OpensergoRuleAggregator {
	return &OpensergoRuleAggregator{
		dataMap:                   make(map[string][]flow.Rule),
		pbTtRuleMapByStrategyKind: make(map[string][]faulttolerancePb.FaultToleranceRule),

		pbRlStrategyMap: make(map[string]faulttolerancePb.RateLimitStrategy),
		ruleAssembler:   RuleAssemblerAggregator{},
	}
}

func (aggregator *OpensergoRuleAggregator) setSentinelUpdateHandler(sentinelUpdateHandler func()) {
	aggregator.sentinelUpdateHandler = sentinelUpdateHandler
}

var updateFaultToleranceRules_Mutex sync.Mutex

func (aggregator *OpensergoRuleAggregator) updateFaultToleranceRules(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	updateFaultToleranceRules_Mutex.Lock()
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
	updateFaultToleranceRules_Mutex.Unlock()
	return true, nil
}

var updateRateLimitStrategy_mutex sync.Mutex

func (aggregator *OpensergoRuleAggregator) updateRateLimitStrategy(dataSlice []protoreflect.ProtoMessage) (bool, error) {
	updateRateLimitStrategy_mutex.Lock()
	if len(dataSlice) > 0 {
		for _, pbData := range dataSlice {
			rateLimitStrategy := pbData.(*faulttolerancePb.RateLimitStrategy)
			aggregator.pbRlStrategyMap[rateLimitStrategy.Name] = *rateLimitStrategy
		}
	}

	aggregator.updateFlowRule()
	updateRateLimitStrategy_mutex.Unlock()
	return true, nil
}

var updateFlowRule_mutex sync.Mutex

// TODO update all flow.Rule now, but in this mode, the performance would be affected when the data becoming large.
func (aggregator *OpensergoRuleAggregator) updateFlowRule() {
	updateFlowRule_mutex.Lock()
	flowRules := make([]flow.Rule, 0)
	pbRuleOfRateLimitStrategies := aggregator.pbTtRuleMapByStrategyKind[configkind.ConfigKindRefRateLimitStrategy{}.GetSimpleName()]
	flowRulesByRlStrategy := aggregator.ruleAssembler.assembleFlowRulesFromRateLimitStrategies(pbRuleOfRateLimitStrategies, aggregator.pbRlStrategyMap)
	if flowRulesByRlStrategy != nil && len(flowRulesByRlStrategy) > 0 {
		flowRules = append(flowRules, flowRulesByRlStrategy...)
	}
	// TODO update
	aggregator.dataMap[RuleType_FlowRule] = flowRules
	aggregator.sentinelUpdateHandler()
	updateFlowRule_mutex.Unlock()
}

func (aggregator *OpensergoRuleAggregator) updateCircuitBreakerRule() {

}
