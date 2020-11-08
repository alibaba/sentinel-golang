package base

import (
	"sync/atomic"
)

const (
	SlotOrderGap = 10000
)

// Default orders of prepare slots.
const (
	ResourceNodePrepareSlotDefaultOrder SlotOrder = SlotOrderGap * (iota + 1)
)

// Default orders of rule check slots.
const (
	SystemAdaptiveSlotDefaultOrder SlotOrder = SlotOrderGap * (iota + 1)
	FlowSlotDefaultOrder
	IsolationSlotDefaultOrder
	CircuitBreakerSlotDefaultOrder
	HotSpotSlotDefaultOrder
)

// Default orders of stat slots.
const (
	StatSlotDefaultOrder SlotOrder = SlotOrderGap * (iota + 1)
	LogSlotDefaultOrder
	CircuitBreakerMetricStatSlotDefaultOrder
	ConcurrencyStatSlotDefaultOrder
	FlowStandaloneStatSlotDefaultOrder
)

// SlotOrder holds order value of slot.
type SlotOrder uint32

func (o *SlotOrder) Order() uint32 {
	return atomic.LoadUint32((*uint32)(o))
}

func (o *SlotOrder) SetOrder(order uint32) {
	atomic.StoreUint32((*uint32)(o), order)
}
