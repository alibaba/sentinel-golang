package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
)

type Slot struct {
}

func (s *Slot) OnEntryPassed(ctx *base.EntryContext) {
	s.recordPassFor(ctx.StatNode, ctx.Input.AcquireCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordPassFor(InboundNode(), ctx.Input.AcquireCount)
	}
}

func (s *Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	s.recordBlockFor(ctx.StatNode, ctx.Input.AcquireCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordBlockFor(InboundNode(), ctx.Input.AcquireCount)
	}
}

func (s *Slot) OnCompleted(ctx *base.EntryContext) {
	rt := util.CurrentTimeMillis() - ctx.StartTime()
	ctx.PutRt(rt)
	s.recordCompleteFor(ctx.StatNode, ctx.Input.AcquireCount, rt, ctx.Err())
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordCompleteFor(InboundNode(), ctx.Input.AcquireCount, rt, ctx.Err())
	}
}

func (s *Slot) recordPassFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.IncreaseGoroutineNum()
	sn.AddMetric(base.MetricEventPass, uint64(count))
}

func (s *Slot) recordBlockFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.AddMetric(base.MetricEventBlock, uint64(count))
}

func (s *Slot) recordCompleteFor(sn base.StatNode, count uint32, rt uint64, err error) {
	if sn == nil {
		return
	}
	if err != nil {
		sn.AddMetric(base.MetricEventError, uint64(count))
	}
	sn.AddMetric(base.MetricEventRt, rt)
	sn.AddMetric(base.MetricEventComplete, uint64(count))
	sn.DecreaseGoroutineNum()
}
