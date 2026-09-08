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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/require"
)

func concurrencyTestController(index int, threshold int64) TrafficShapingController {
	return &rejectTrafficShapingController{baseTrafficShapingController: *newBaseTrafficShapingController(&Rule{
		Resource: "concurrency-regression", MetricType: Concurrency, ParamIndex: index, Threshold: threshold,
		SpecificItems: map[interface{}]int64{"special": 2},
	})}
}

func installConcurrencyTestControllers(t *testing.T, controllers ...TrafficShapingController) {
	t.Helper()
	tcMux.Lock()
	previous := tcMap
	tcMap = trafficControllerMap{"concurrency-regression": controllers}
	tcMux.Unlock()
	t.Cleanup(func() { tcMux.Lock(); tcMap = previous; tcMux.Unlock() })
}

func concurrencyTestContext(args ...interface{}) *base.EntryContext {
	return &base.EntryContext{
		Resource:        base.NewResourceWrapper("concurrency-regression", base.ResTypeCommon, base.Outbound),
		Input:           &base.SentinelInput{BatchCount: 1, Args: args},
		RuleCheckResult: base.NewTokenResultPass(),
	}
}

func requireConcurrencyCounter(t *testing.T, tc TrafficShapingController, arg interface{}, expected int64) {
	t.Helper()
	counter, ok := tc.BoundMetric().ConcurrencyCounter.Get(arg)
	require.True(t, ok)
	require.NotNil(t, counter)
	require.Equal(t, expected, atomic.LoadInt64(counter))
}

// All workers finish checking before any admitted entry completes. This forces
// contention and measures accepted in-flight entries independently of counters.
func TestConcurrencySlotConcurrentLimit(t *testing.T) {
	for _, arg := range []string{"ordinary", "special"} {
		t.Run(arg, func(t *testing.T) {
			controller := concurrencyTestController(0, 5)
			installConcurrencyTestControllers(t, controller)
			chain := base.NewSlotChain()
			chain.AddRuleCheckSlot(&Slot{})
			chain.AddStatSlot(&ConcurrencyStatSlot{})
			threshold := int64(5)
			if arg == "special" {
				threshold = 2
			}
			const workers = 64
			for round := 0; round < 20; round++ {
				start, release := make(chan struct{}), make(chan struct{})
				var checked, done sync.WaitGroup
				checked.Add(workers)
				done.Add(workers)
				var active, exceeded, blocked, invalid int64
				for i := 0; i < workers; i++ {
					go func() {
						defer done.Done()
						ctx := concurrencyTestContext(arg)
						<-start
						result := chain.Entry(ctx)
						passed := result != nil && !result.IsBlocked()
						if result == nil {
							atomic.AddInt64(&invalid, 1)
						}
						if passed {
							if atomic.AddInt64(&active, 1) > threshold {
								atomic.AddInt64(&exceeded, 1)
							}
						} else {
							atomic.AddInt64(&blocked, 1)
						}
						checked.Done()
						<-release
						if passed {
							atomic.AddInt64(&active, -1)
							(&ConcurrencyStatSlot{}).OnCompleted(ctx)
						}
					}()
				}
				close(start)
				checked.Wait()
				// Release workers before assertions that may terminate the test.
				admitted := atomic.LoadInt64(&active)
				counter, ok := controller.BoundMetric().ConcurrencyCounter.Get(arg)
				var reserved int64
				if ok && counter != nil {
					reserved = atomic.LoadInt64(counter)
				}
				close(release)
				done.Wait()
				require.Zero(t, invalid)
				require.Zero(t, exceeded)
				// Provisional increments from rejected contenders may reduce admissions.
				require.Positive(t, admitted)
				require.LessOrEqual(t, admitted, threshold)
				require.Equal(t, int64(workers)-admitted, blocked)
				require.True(t, ok)
				require.NotNil(t, counter)
				require.Equal(t, admitted, reserved)
				require.Zero(t, active)
				requireConcurrencyCounter(t, controller, arg, 0)
			}
		})
	}
}

type mutatingConcurrencyCheckSlot struct{ block bool }

func (s *mutatingConcurrencyCheckSlot) Order() uint32 { return RuleCheckSlotOrder + 1 }
func (s *mutatingConcurrencyCheckSlot) Check(ctx *base.EntryContext) *base.TokenResult {
	ctx.Input.Args[0] = "changed"
	ctx.Input.Attachments["key"] = "changed"
	if s.block {
		return base.NewTokenResultBlocked(base.BlockTypeFlow)
	}
	return nil
}

func TestConcurrencySlotConcurrentRollback(t *testing.T) {
	for _, mode := range []string{"in-slot", "later-slot", "completion"} {
		t.Run(mode, func(t *testing.T) {
			const workers = 64
			first, second := concurrencyTestController(0, workers), concurrencyTestController(1, workers)
			second.BoundRule().ParamKey = "key"
			second.(*rejectTrafficShapingController).paramKey = "key"
			skipped := concurrencyTestController(2, workers)
			controllers := []TrafficShapingController{first, second, skipped}
			if mode == "in-slot" {
				controllers = append(controllers, concurrencyTestController(0, 0))
			}
			installConcurrencyTestControllers(t, controllers...)
			// Decoy counters must never be decremented by re-extracting mutated input.
			for _, tc := range controllers {
				value := int64(7)
				tc.BoundMetric().ConcurrencyCounter.Add("changed", &value)
			}
			chain := base.NewSlotChain()
			chain.AddRuleCheckSlot(&Slot{})
			if mode != "in-slot" {
				chain.AddRuleCheckSlot(&mutatingConcurrencyCheckSlot{block: mode == "later-slot"})
			}
			chain.AddStatSlot(&ConcurrencyStatSlot{})
			start := make(chan struct{})
			var done sync.WaitGroup
			var incorrect int64
			done.Add(workers)
			for i := 0; i < workers; i++ {
				go func() {
					defer done.Done()
					ctx := concurrencyTestContext("original", "fallback")
					ctx.Input.Attachments = map[interface{}]interface{}{"key": "attached"}
					<-start
					result := chain.Entry(ctx)
					if result == nil || result.IsBlocked() != (mode != "completion") {
						atomic.AddInt64(&incorrect, 1)
					}
					if mode == "completion" {
						(&ConcurrencyStatSlot{}).OnCompleted(ctx)
					}
					// Releasing consumed bookkeeping must not decrement counters twice.
					(&ConcurrencyStatSlot{}).OnEntryBlocked(ctx, nil)
					(&ConcurrencyStatSlot{}).OnCompleted(ctx)
					if _, ok := ctx.Data[hotspotConcurrencyPassedKey]; ok {
						atomic.AddInt64(&incorrect, 1)
					}
				}()
			}
			close(start)
			done.Wait()
			require.Zero(t, incorrect)
			requireConcurrencyCounter(t, first, "original", 0)
			requireConcurrencyCounter(t, second, "attached", 0)
			for _, tc := range controllers {
				requireConcurrencyCounter(t, tc, "changed", 7)
			}
			require.Equal(t, 1, skipped.BoundMetric().ConcurrencyCounter.Len())
			if mode == "in-slot" {
				requireConcurrencyCounter(t, controllers[3], "original", 0)
			}
		})
	}
}

func TestConcurrencySlotReleaseAfterControllersChange(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		t.Run(map[bool]string{false: "completion", true: "blocked"}[blocked], func(t *testing.T) {
			original := concurrencyTestController(0, 1)
			installConcurrencyTestControllers(t, original)
			ctx := concurrencyTestContext("original")
			require.False(t, (&Slot{}).Check(ctx).IsBlocked())
			requireConcurrencyCounter(t, original, "original", 1)
			replacement := concurrencyTestController(0, 1)
			value := int64(1)
			replacement.BoundMetric().ConcurrencyCounter.Add("original", &value)
			tcMux.Lock()
			tcMap["concurrency-regression"] = []TrafficShapingController{replacement}
			tcMux.Unlock()
			ctx.Resource = nil
			ctx.Input = nil
			if blocked {
				(&ConcurrencyStatSlot{}).OnEntryBlocked(ctx, nil)
			} else {
				(&ConcurrencyStatSlot{}).OnCompleted(ctx)
			}
			requireConcurrencyCounter(t, original, "original", 0)
			requireConcurrencyCounter(t, replacement, "original", 1)
		})
	}
}

func TestConcurrencySlotCleanupWithoutReservations(t *testing.T) {
	for _, ctx := range []*base.EntryContext{
		nil,
		{},
		{Data: map[interface{}]interface{}{"unrelated": "keep"}},
	} {
		slot := &ConcurrencyStatSlot{}
		require.NotPanics(t, func() {
			slot.OnEntryBlocked(ctx, nil)
			slot.OnCompleted(ctx)
			slot.OnEntryBlocked(ctx, nil)
		})
		if ctx != nil && ctx.Data != nil {
			require.Equal(t, map[interface{}]interface{}{"unrelated": "keep"}, ctx.Data)
		}
	}
}
