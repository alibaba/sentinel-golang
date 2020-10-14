package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/log"
	"github.com/alibaba/sentinel-golang/core/misc"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/core/system"
)

var globalSlotChain = BuildDefaultSlotChain()

func GlobalSlotChain() *base.SlotChain {
	return globalSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(stat.DefaultResourceNodePrepareSlot)

	sc.AddRuleCheckSlotLast(system.DefaultAdaptiveSlot)
	sc.AddRuleCheckSlotLast(flow.DefaultSlot)
	sc.AddRuleCheckSlotLast(isolation.DefaultSlot)
	sc.AddRuleCheckSlotLast(circuitbreaker.DefaultSlot)
	sc.AddRuleCheckSlotLast(hotspot.DefaultSlot)

	sc.AddStatSlotLast(stat.DefaultSlot)
	sc.AddStatSlotLast(log.DefaultSlot)
	sc.AddStatSlotLast(circuitbreaker.DefaultMetricStatSlot)
	sc.AddStatSlotLast(hotspot.DefaultConcurrencyStatSlot)
	sc.AddStatSlotLast(flow.DefaultStandaloneStatSlot)

	misc.RegisterCustomGlobalSlotsToSc(sc)
	return sc
}
