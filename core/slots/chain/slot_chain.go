package chain

import (
	"container/list"
	"context"
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"github.com/sentinel-group/sentinel-golang/core/slots/flow"
	"github.com/sentinel-group/sentinel-golang/core/slots/statistic"
)

type SlotChain interface {
	Slot
	AddFirst(slot Slot)
	AddLast(slot Slot)
}

type DefaultSlotChain struct {
	slots *list.List
}

type SlotChainBuilder interface {
	Build() SlotChain
}

func NewDefaultSlotChain() *DefaultSlotChain {
	defaultSlotChain := &DefaultSlotChain{
		slots: list.New(),
	}
	defaultSlotChain.AddLast(&flow.FlowSlot{
		RuleManager: flow.NewRuleManager(),
	})
	defaultSlotChain.AddLast(&statistic.StatisticSlot{})
	return defaultSlotChain
}

func (dsc *DefaultSlotChain) AddFirst(slot Slot) {
	dsc.slots.PushFront(slot)
}

func (dsc *DefaultSlotChain) AddLast(slot Slot) {
	dsc.slots.PushBack(slot)
}

func (dsc *DefaultSlotChain) IsContinue(lastResult base.SlotResult, ctx context.Context) bool {
	return true
}

func (dsc *DefaultSlotChain) Entry(ctx context.Context, resourceWrap *base.ResourceWrapper, node *base.DefaultNode, count uint32) base.SlotResult {
	slotResult := base.SlotResult{
		Status: base.ResultStatusOk,
	}
	for e := dsc.slots.Front(); e != nil; e = e.Next() {
		slot := e.Value
		switch slot_ := slot.(type) {
		case *flow.FlowSlot:
			if slot_.IsContinue(slotResult, ctx) {
				slotResult = slot_.Entry(ctx, resourceWrap, node, count)
			}
			break
		case *statistic.StatisticSlot:
			if slot_.IsContinue(slotResult, ctx) {
				slotResult = slot_.Entry(ctx, resourceWrap, node, count)
			}
			break
		default:
			slotResult = base.SlotResult{
				Status:        base.ResultStatusBlocked,
				BlockedReason: "Unknown Slot",
			}
		}
	}
	return slotResult
}

func (dsc *DefaultSlotChain) Exit(ctx context.Context, resourceWrap *base.ResourceWrapper, count uint32) {
	for e := dsc.slots.Front(); e != nil; e = e.Next() {
		slot := e.Value
		switch slot_ := slot.(type) {
		case *flow.FlowSlot:
			slot_.Exit(ctx, resourceWrap, count)
			break
		case *statistic.StatisticSlot:
			slot_.Exit(ctx, resourceWrap, count)
			break
		default:
			fmt.Println("DefaultSlotChain Exit error!")
		}
	}
}
