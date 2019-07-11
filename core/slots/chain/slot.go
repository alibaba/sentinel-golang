package chain

import (
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

// a solt
type Slot interface {
	/**
	 * Entrance of this slots.
	 */
	Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error)

	Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error

	// 传递进入
	FireEntry(context *base.Context, resourceWrapper *base.ResourceWrapper, defaultNode *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error)

	// 传递退出
	FireExit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error

	GetNext() Slot

	SetNext(next Slot)
}

// a slot can make slot compose linked
type LinkedSlot struct {
	// next linkedSlot
	next Slot
}

// 传递退出
func (s *LinkedSlot) Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	return s.FireEntry(ctx, resWrapper, node, count, prioritized)
}

// 传递进入
func (s *LinkedSlot) Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	return s.FireExit(context, resourceWrapper, count)
}

// 传递进入, 没有下一个就返回 ResultStatusPass
func (s *LinkedSlot) FireEntry(context *base.Context, resourceWrapper *base.ResourceWrapper, defaultNode *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	if s.next != nil {
		return s.next.Entry(context, resourceWrapper, defaultNode, count, prioritized)
	}
	return base.NewSlotResultPass(), nil
}

// 传递退出，没有下一个就返回
func (s *LinkedSlot) FireExit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	if s.next != nil {
		return s.next.Exit(context, resourceWrapper, count)
	} else {
		return nil
	}
}

func (s *LinkedSlot) GetNext() Slot {
	return s.next
}

func (s *LinkedSlot) SetNext(next Slot) {
	s.next = next
}
