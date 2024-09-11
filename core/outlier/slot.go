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

package outlier

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

const (
	RuleCheckSlotOrder = 6000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(resource) == 0 {
		return result
	}
	filterNodes, outlierNodes, halfOpenNodes := checkAllNodes(ctx)
	if len(filterNodes) != 0 {
		result.SetFilterNodes(filterNodes)
	}
	if len(halfOpenNodes) != 0 {
		result.SetHalfOpenNodes(halfOpenNodes)
	}
	if len(outlierNodes) != 0 {
		retryer := getRetryerOfResource(resource)
		retryer.scheduleRetry(outlierNodes)
		recycler := getRecyclerOfResource(resource)
		recycler.scheduleRecycler(outlierNodes)
	}
	return result
}

func checkAllNodes(ctx *base.EntryContext) (filters []string, outliers []string, halfs []string) {
	resource := ctx.Resource.Name()
	nodeBreaks := getNodeBreakerOfResource(resource)
	rule := getOutlierRuleOfResource(resource)
	nodeCount := getNodeCountOfResource(resource)
	for nodeID, breaker := range nodeBreaks {
		if breaker.TryPass(ctx) {
			if !rule.EnableActiveRecovery && breaker.CurrentState() == circuitbreaker.HalfOpen {
				halfs = append(halfs, nodeID)
			}
			continue
		}
		if rule.EnableActiveRecovery {
			outliers = append(outliers, nodeID)
		}
		if len(filters) <= int(float64(nodeCount)*rule.MaxEjectionPercent) {
			filters = append(filters, nodeID)
		}
	}
	return filters, outliers, halfs
}
