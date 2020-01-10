package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
)

const SlotName = "StatisticSlot"

type StatisticSlot struct {
}

func (s *StatisticSlot) String() string {
	return SlotName
}

func (s *StatisticSlot) OnEntryPassed(ctx *base.EntryContext) {
	s.recordPassFor(ctx.StatNode, ctx.Input.AcquireCount)
}

func (s *StatisticSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	s.recordBlockFor(ctx.StatNode, ctx.Input.AcquireCount)
}

func (s *StatisticSlot) OnCompleted(ctx *base.EntryContext) {
	if ctx.Output.LastResult == nil || ctx.Output.LastResult.IsBlocked() {
		return
	}
	rt := util.CurrentTimeMillis() - ctx.StartTime()
	s.recordCompleteFor(ctx.StatNode, ctx.Input.AcquireCount, rt)
}

func (s *StatisticSlot) recordPassFor(sn base.StatNode, count uint32) {
	logger.Debug("Entry passed.")
	if sn == nil {
		return
	}
	sn.IncreaseGoroutineNum()
	sn.AddRequest(base.MetricEventPass, uint64(count))
}

func (s *StatisticSlot) recordBlockFor(sn base.StatNode, count uint32) {
	logger.Debug("Entry blocked.")
	if sn == nil {
		return
	}
	sn.AddRequest(base.MetricEventBlock, uint64(count))
}

func (s *StatisticSlot) recordCompleteFor(sn base.StatNode, count uint32, rt uint64) {
	logger.Debug("Entry completed.")
	if sn == nil {
		return
	}
	sn.AddRtAndCompleteRequest(rt, uint64(count))
	sn.DecreaseGoroutineNum()
}
