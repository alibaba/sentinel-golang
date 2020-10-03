package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
)

type Slot struct {
}

func (s *Slot) OnEntryPassed(ctx *base.EntryContext) {
	s.recordPassFor(ctx.StatNode, ctx.Input.BatchCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordPassFor(InboundNode(), ctx.Input.BatchCount)
	}
}

func (s *Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	s.recordBlockFor(ctx.StatNode, ctx.Input.BatchCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordBlockFor(InboundNode(), ctx.Input.BatchCount)
	}
}

func (s *Slot) OnCompleted(ctx *base.EntryContext) {
	rt := util.CurrentTimeMillis() - ctx.StartTime()
	ctx.PutRt(rt)
	s.recordCompleteFor(ctx.StatNode, ctx.Input.BatchCount, rt, ctx.Err())
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordCompleteFor(InboundNode(), ctx.Input.BatchCount, rt, ctx.Err())
	}
}

func (s *Slot) recordPassFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.IncreaseGoroutineNum()
	sn.AddCount(base.MetricEventPass, int64(count))
}

func (s *Slot) recordBlockFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.AddCount(base.MetricEventBlock, int64(count))
}

func (s *Slot) recordCompleteFor(sn base.StatNode, count uint32, rt uint64, err error) {
	if sn == nil {
		return
	}
	if err != nil {
		sn.AddCount(base.MetricEventError, int64(count))
	}
	sn.AddCount(base.MetricEventRt, int64(rt))
	sn.AddCount(base.MetricEventComplete, int64(count))
	sn.DecreaseGoroutineNum()
}
