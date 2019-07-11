package chain

import (
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type SlotChain interface {
	/**
	 * Add a processor to the head of this slots slotchain.
	 *
	 * @param protocolProcessor processor to be added.
	 */
	AddFirst(slot Slot)

	/**
	 * Add a processor to the tail of this slots slotchain.
	 *
	 * @param protocolProcessor processor to be added.
	 */
	AddLast(slot Slot)

	// fire to next slot
	Entry(context *base.Context, resourceWrapper *base.ResourceWrapper, defaultNode *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error)
	// fire to next slot
	Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error
}

// implent SlotChain
type LinkedSlotChain struct {
	first Slot
	end   Slot
}

func NewLinkedSlotChain() *LinkedSlotChain {
	fs := new(LinkedSlot)
	return &LinkedSlotChain{first: fs, end: fs}
}

func (lsc *LinkedSlotChain) AddFirst(slot Slot) {
	slot.SetNext(lsc.first.GetNext())
	lsc.first.SetNext(slot)
	if lsc.end == lsc.first {
		lsc.end = slot
	}
}

func (lsc *LinkedSlotChain) AddLast(slot Slot) {
	lsc.end.SetNext(slot)
	lsc.end = slot
}

func (lsc *LinkedSlotChain) Entry(context *base.Context, resourceWrapper *base.ResourceWrapper, defaultNode *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	return lsc.first.Entry(context, resourceWrapper, defaultNode, count, prioritized)
}

// 传递进入
func (lsc *LinkedSlotChain) Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	return lsc.first.Exit(context, resourceWrapper, count)
}

type SlotChainBuilder interface {
	Build() SlotChain
}
