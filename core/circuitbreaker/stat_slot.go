package circuitbreaker

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

const (
	StatSlotName  = "sentinel-core-circuit-breaker-metric-stat-slot"
	StatSlotOrder = 3000
)

var (
	DefaultMetricStatSlot = &MetricStatSlot{}
)

// MetricStatSlot records metrics for circuit breaker on invocation completed.
// MetricStatSlot must be filled into slot chain if circuit breaker is alive.
type MetricStatSlot struct {
}

func (s *MetricStatSlot) Name() string {
	return StatSlotName
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
	for _, cb := range getBreakersOfResource(res) {
		cb.OnRequestComplete(rt, err)
	}
}
