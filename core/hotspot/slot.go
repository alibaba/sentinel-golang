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
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
)

const (
	RuleCheckSlotOrder = 4000
)

// hotspotConcurrencyPassedKey stores the exact reservations made in the Check phase.
const hotspotConcurrencyPassedKey = "sentinel_hotspot_concurrency_passed"

// passedConcurrencyEntry identifies the counter to release, even if the context or
// resource rules change after checking.
type passedConcurrencyEntry struct {
	tc  TrafficShapingController
	arg interface{}
}

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	batch := int64(ctx.Input.BatchCount)

	result := ctx.RuleCheckResult
	tcs := getTrafficControllersFor(res)

	// Track concurrency controllers that passed (and incremented their counters)
	// so we can rollback if a later controller in the same slot blocks.
	var passedConcurrency []passedConcurrencyEntry

	for _, tc := range tcs {
		arg := tc.ExtractArgs(ctx)
		if arg == nil {
			continue
		}
		r := canPassCheck(tc, arg, batch)
		if r == nil {
			if tc.BoundRule().MetricType == Concurrency {
				if passedConcurrency == nil {
					passedConcurrency = make([]passedConcurrencyEntry, 0, len(tcs))
				}
				passedConcurrency = append(passedConcurrency, passedConcurrencyEntry{tc, arg})
			}
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			// Rollback concurrency counters incremented by earlier controllers in this slot.
			releaseConcurrency(passedConcurrency)
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
				// Handle waiting action.
				util.Sleep(nanosToWait)
			}
			continue
		}
	}

	// Retain reservations for completion or rollback if a later slot blocks.
	if len(passedConcurrency) > 0 {
		if ctx.Data == nil {
			ctx.Data = make(map[interface{}]interface{})
		}
		ctx.Data[hotspotConcurrencyPassedKey] = passedConcurrency
	}

	return result
}

func canPassCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return canPassLocalCheck(tc, arg, batch)
}

func canPassLocalCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return tc.PerformChecking(arg, batch)
}
