/**
 * @description:
 *
 * @author: helloworld
 * @date:2019-07-11
 */
package chain

import (
	"errors"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"testing"
)

// implent slot for unit test

// 继承 LinkedProessorSlot 并完全实现 Slot
type IncrSlot struct {
	LinkedSlot
}

func (s *IncrSlot) Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	count++
	return s.FireEntry(ctx, resWrapper, node, count, prioritized)
}

func (s *IncrSlot) Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	count++
	return s.FireExit(context, resourceWrapper, count)
}

// 继承 LinkedProessorSlot 并完全实现 Slot
type DecrSlot struct {
	LinkedSlot
}

func (s *DecrSlot) Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	count--
	return s.FireEntry(ctx, resWrapper, node, count, prioritized)
}

func (s *DecrSlot) Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	count--
	return s.FireExit(context, resourceWrapper, count)
}

// 继承 LinkedProessorSlot 并完全实现 Slot
type GreaterZeroPassSlot struct {
	num int
	LinkedSlot
}

func (s *GreaterZeroPassSlot) Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	if count > s.num {
		return s.FireEntry(ctx, resWrapper, node, count, prioritized)
	} else {
		return base.NewSlotResultBlock("GreaterZeroPassSlot"), nil
	}
}

func (s *GreaterZeroPassSlot) Exit(context *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	if count <= 0 {
		return errors.New("GreaterZeroPassSlot")
	}
	return s.FireExit(context, resourceWrapper, count)
}

func TestLinkedSlotChain_AddFirst_Pass(t *testing.T) {
	newChain := NewLinkedSlotChain()
	newChain.AddFirst(new(GreaterZeroPassSlot))
	newChain.AddFirst(new(IncrSlot))

	result, _ := newChain.Entry(nil, nil, nil, 0, false)
	if result.Status != base.ResultStatusPass {
		t.Fatal("TestLinkedSlotChain_AddFirst_Block")
	}
	err := newChain.Exit(nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLinkedSlotChain_AddFirst_Block(t *testing.T) {
	newChain := NewLinkedSlotChain()
	newChain.AddFirst(new(IncrSlot))
	newChain.AddFirst(new(GreaterZeroPassSlot))
	newChain.AddFirst(new(DecrSlot))

	result, _ := newChain.Entry(nil, nil, nil, 0, false)
	if result.Status != base.ResultStatusBlocked {
		t.Fatal("TestLinkedSlotChain_AddFirst_Block")
	}
	err := newChain.Exit(nil, nil, 0)
	if err == nil {
		t.Fatal("should has error")
	}
}

func TestLinkedSlotChain_AddLast_Pass(t *testing.T) {
	newChain := NewLinkedSlotChain()
	newChain.AddLast(new(IncrSlot))
	newChain.AddLast(new(IncrSlot))
	newChain.AddLast(new(DecrSlot))
	newChain.AddLast(new(GreaterZeroPassSlot))
	result, _ := newChain.Entry(nil, nil, nil, 0, false)
	if result.Status != base.ResultStatusPass {
		t.Fatal("TestLinkedSlotChain_AddLast_Pass")
	}

	err := newChain.Exit(nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLinkedSlotChain_AddLast_Block(t *testing.T) {
	newChain := NewLinkedSlotChain()
	newChain.AddLast(new(IncrSlot))
	newChain.AddLast(new(DecrSlot))
	newChain.AddLast(new(DecrSlot))
	newChain.AddLast(new(GreaterZeroPassSlot))

	result, _ := newChain.Entry(nil, nil, nil, 0, false)
	if result.Status != base.ResultStatusBlocked {
		t.Fatal("TestLinkedSlotChain_AddLast_Block")
	}

	err := newChain.Exit(nil, nil, 0)
	if err == nil {
		t.Fatal("should has error")
	}
}
