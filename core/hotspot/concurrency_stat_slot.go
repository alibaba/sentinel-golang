package hotspot

import (
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	StatSlotName  = "sentinel-core-hotspot-concurrency-stat-slot"
	StatSlotOrder = 4000
)

var (
	DefaultConcurrencyStatSlot = &ConcurrencyStatSlot{}
)

// ConcurrencyStatSlot is to record the Concurrency statistic for all arguments
type ConcurrencyStatSlot struct {
}

func (s *ConcurrencyStatSlot) Name() string {
	return StatSlotName
}

func (s *ConcurrencyStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (c *ConcurrencyStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	args := ctx.Input.Args
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		arg := matchArg(tc, args)
		if arg == nil {
			continue
		}
		metric := tc.BoundMetric()
		concurrencyPtr, existed := metric.ConcurrencyCounter.Get(arg)
		if !existed || concurrencyPtr == nil {
			if logging.DebugEnabled() {
				logging.Debug("[ConcurrencyStatSlot OnEntryPassed] Parameter does not exist in ConcurrencyCounter.", "argument", arg)
			}
			continue
		}
		atomic.AddInt64(concurrencyPtr, 1)
	}
}

func (c *ConcurrencyStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// Do nothing
}

func (c *ConcurrencyStatSlot) OnCompleted(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	args := ctx.Input.Args
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		arg := matchArg(tc, args)
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
