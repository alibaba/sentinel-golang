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

package hotspot

import (
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	StatSlotOrder = 4000
)

var (
	DefaultConcurrencyStatSlot = &ConcurrencyStatSlot{}
)

// ConcurrencyStatSlot is to record the Concurrency statistic for all arguments
type ConcurrencyStatSlot struct {
}

func (s *ConcurrencyStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (c *ConcurrencyStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	// The concurrency counter is already incremented atomically in the Check phase
	// (performCheckingForConcurrencyMetric), so no increment is needed here.
}

func (c *ConcurrencyStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// If hotspot concurrency counters were incremented in the Check phase but the
	// entry was blocked by a later check slot (e.g., CircuitBreaker), we need to
	// rollback those increments.
	if ctx.Data == nil {
		return
	}
	if _, ok := ctx.Data[hotspotConcurrencyPassedKey]; !ok {
		return
	}
	res := ctx.Resource.Name()
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		if tc.BoundRule().MetricType != Concurrency {
			continue
		}
		arg := tc.ExtractArgs(ctx)
		if arg == nil {
			continue
		}
		metric := tc.BoundMetric()
		concurrencyPtr, existed := metric.ConcurrencyCounter.Get(arg)
		if !existed || concurrencyPtr == nil {
			continue
		}
		atomic.AddInt64(concurrencyPtr, -1)
	}
}

func (c *ConcurrencyStatSlot) OnCompleted(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		if tc.BoundRule().MetricType != Concurrency {
			continue
		}
		arg := tc.ExtractArgs(ctx)
		if arg == nil {
			continue
		}
		metric := tc.BoundMetric()
		concurrencyPtr, existed := metric.ConcurrencyCounter.Get(arg)
		if !existed || concurrencyPtr == nil {
			if logging.DebugEnabled() {
				logging.Debug("[ConcurrencyStatSlot OnCompleted] Parameter does not exist in ConcurrencyCounter.", "argument", arg)
			}
			continue
		}
		atomic.AddInt64(concurrencyPtr, -1)
	}
}
