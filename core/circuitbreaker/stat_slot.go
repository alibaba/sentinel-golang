package circuitbreaker

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

// MetricStatSlot records metrics for circuit breaker on invocation completed.
// MetricStatSlot must be filled into slot chain if circuit breaker is alive.
type MetricStatSlot struct {
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
	for _, cb := range getResBreakers(res) {
		cb.OnRequestComplete(rt, err)
	}
}
