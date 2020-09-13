package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/log"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/core/system"
)

var globalSlotChain = BuildDefaultSlotChain()

// SetSlotChain replaces current slot chain with the given one.
// Note that this operation is not thread-safe, so it should be
// called when pre-initializing Sentinel.
func SetSlotChain(chain *base.SlotChain) {
	if chain != nil {
		globalSlotChain = chain
	}
}

func GlobalSlotChain() *base.SlotChain {
	return globalSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(&stat.ResourceNodePrepareSlot{})
	sc.AddRuleCheckSlotLast(&system.AdaptiveSlot{})
	sc.AddRuleCheckSlotLast(&flow.Slot{})
	sc.AddRuleCheckSlotLast(&circuitbreaker.Slot{})
	sc.AddRuleCheckSlotLast(&hotspot.Slot{})
	sc.AddStatSlotLast(&stat.Slot{})
	sc.AddStatSlotLast(&log.Slot{})
	sc.AddStatSlotLast(&circuitbreaker.MetricStatSlot{})
	sc.AddStatSlotLast(&hotspot.ConcurrencyStatSlot{})
	return sc
}
