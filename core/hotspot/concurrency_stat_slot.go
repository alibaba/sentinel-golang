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
	releaseContextConcurrency(ctx)
}

func (c *ConcurrencyStatSlot) OnCompleted(ctx *base.EntryContext) {
	releaseContextConcurrency(ctx)
}

func releaseContextConcurrency(ctx *base.EntryContext) {
	if ctx == nil {
		return
	}
	passed, _ := ctx.Data[hotspotConcurrencyPassedKey].([]passedConcurrencyEntry)
	// Consume reservations before releasing them so repeated callbacks are harmless.
	delete(ctx.Data, hotspotConcurrencyPassedKey)
	releaseConcurrency(passed)
}

func releaseConcurrency(passed []passedConcurrencyEntry) {
	for _, p := range passed {
		counter, existed := p.tc.BoundMetric().ConcurrencyCounter.Get(p.arg)
		if existed && counter != nil {
			atomic.AddInt64(counter, -1)
		}
	}
}
