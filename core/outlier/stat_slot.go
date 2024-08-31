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
	StatSlotOrder = 6000
)

var (
	DefaultMetricStatSlot = &MetricStatSlot{}
)

// MetricStatSlot records metrics for outlier ejection on invocation completed.
// MetricStatSlot must be filled into slot chain if outlier ejection is alive.
type MetricStatSlot struct {
}

func (s *MetricStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (c *MetricStatSlot) OnEntryPassed(_ *base.EntryContext) {
	// Do nothing
	return
}

func (c *MetricStatSlot) OnEntryBlocked(_ *base.EntryContext, _ *base.BlockError) {
	// Do nothing
	return
}

func (c *MetricStatSlot) OnCompleted(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	err := ctx.Err()
	rt := ctx.Rt()
	nodes := getNodeBreakersOfResource(res)
	address := ctx.GetPair("address").(string)
	if address == "" {
		return
	}
	if _, ok := nodes[address]; !ok {
		newOneBreakers(res, address)
		nodes = getNodeBreakersOfResource(res)
	}
	for _, breaker := range nodes[address] {
		breaker.OnRequestComplete(rt, err)
	}
}

func newOneBreakers(res string, address string) {
	old := make([]circuitbreaker.CircuitBreaker, 0)
	newCbsOfRes := circuitbreaker.BuildResourceCircuitBreaker(res, breakerRules[res], old)
	updateMux.Lock()
	if len(newCbsOfRes) > 0 {
		if nodeBreakers[res] == nil {
			nodeBreakers[res] = make(map[string][]circuitbreaker.CircuitBreaker)
			nodeCount[res] = 0
		}
		nodeBreakers[res][address] = newCbsOfRes
		nodeCount[res]++
	}
	updateMux.Unlock()
}
