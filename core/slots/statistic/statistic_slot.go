package statistic

import (
	"context"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type StatisticSlot struct {
}

func (ss *StatisticSlot) IsContinue(lastResult base.SlotResult, ctx context.Context) bool {
	return true
}

func (ss *StatisticSlot) Entry(ctx context.Context, resourceWrap *base.ResourceWrapper, node *base.DefaultNode, count uint32) base.SlotResult {
	node.AddGoroutineNum(count)
	node.AddPass(uint64(count))
	return base.SlotResult{
		Status: base.ResultStatusOk,
	}
}

func (ss *StatisticSlot) Exit(ctx context.Context, resourceWrap *base.ResourceWrapper, count uint32) {

}
