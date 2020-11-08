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
	sc.InsertStatPrepareSlotByOrder(stat.DefaultResourceNodePrepareSlot)

	sc.InsertRuleCheckSlotByOrder(system.DefaultAdaptiveSlot)
	sc.InsertRuleCheckSlotByOrder(flow.DefaultSlot)
	sc.InsertRuleCheckSlotByOrder(isolation.DefaultSlot)
	sc.InsertRuleCheckSlotByOrder(circuitbreaker.DefaultSlot)
	sc.InsertRuleCheckSlotByOrder(hotspot.DefaultSlot)

	sc.InsertStatSlotByOrder(stat.DefaultSlot)
	sc.InsertStatSlotByOrder(log.DefaultSlot)
	sc.InsertStatSlotByOrder(circuitbreaker.DefaultMetricStatSlot)
	sc.InsertStatSlotByOrder(hotspot.DefaultConcurrencyStatSlot)
	sc.InsertStatSlotByOrder(flow.DefaultStandaloneStatSlot)
	return sc
}

// RegisterGlobalStatPrepareSlot registers the global StatPrepareSlot for all resource
// Note: this function is not thread-safe
func RegisterGlobalStatPrepareSlot(slot base.StatPrepareSlot) {
	misc.RegisterGlobalStatPrepareSlot(slot)
}

// RegisterGlobalRuleCheckSlot registers the global RuleCheckSlot for all resource
// Note: this function is not thread-safe
func RegisterGlobalRuleCheckSlot(slot base.RuleCheckSlot) {
	misc.RegisterGlobalRuleCheckSlot(slot)
}

// RegisterGlobalStatSlot registers the global StatSlot for all resource
// Note: this function is not thread-safe
func RegisterGlobalStatSlot(slot base.StatSlot) {
	misc.RegisterGlobalStatSlot(slot)
}
