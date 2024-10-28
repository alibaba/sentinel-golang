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
	result.SetFilterNodes(filterNodes)
	result.SetHalfOpenNodes(halfOpenNodes)

	if len(outlierNodes) != 0 {
		rule := getOutlierRuleOfResource(resource)
		if rule.EnableActiveRecovery && len(retryerCh) < capacity {
			retryerCh <- task{outlierNodes, resource}
		}
		if len(recyclerCh) < capacity {
			recyclerCh <- task{outlierNodes, resource}
		}
	}
	return result
}

func checkAllNodes(ctx *base.EntryContext) (filters []string, outliers []string, halfs []string) {
	resource := ctx.Resource.Name()
	nodeBreaks := getNodeBreakersOfResource(resource)
	rule := getOutlierRuleOfResource(resource)
	nodeCount := len(nodeBreaks)
	for address, breaker := range nodeBreaks {
		if breaker.TryPass(ctx) {
			if !rule.EnableActiveRecovery && breaker.CurrentState() == circuitbreaker.HalfOpen {
				halfs = append(halfs, address)
			}
			continue
		}
		outliers = append(outliers, address)
		if len(filters) < int(float64(nodeCount)*rule.MaxEjectionPercent) {
			filters = append(filters, address)
		}
	}
	return filters, outliers, halfs
}
