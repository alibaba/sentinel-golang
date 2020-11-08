package misc

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/log"
	"github.com/alibaba/sentinel-golang/core/stat"
)

var (
	rsSlotChainLock sync.RWMutex
	rsSlotChain     = make(map[string]*base.SlotChain, 8)

	globalStatPrepareSlots = make([]base.StatPrepareSlot, 0, 8)
	globalRuleCheckSlots   = make([]base.RuleCheckSlot, 0, 8)
	globalStatSlot         = make([]base.StatSlot, 0, 8)
)

func registerCustomGlobalSlotsToSc(sc *base.SlotChain) {
	if sc == nil {
		return
	}
	for _, s := range globalStatPrepareSlots {
		if base.ValidateStatPrepareSlotNaming(sc, s) {
			sc.InsertStatPrepareSlotByOrder(s)
		}
	}
	for _, s := range globalRuleCheckSlots {
		if base.ValidateRuleCheckSlotNaming(sc, s) {
			sc.InsertRuleCheckSlotByOrder(s)
		}
	}
	for _, s := range globalStatSlot {
		if base.ValidateStatSlotNaming(sc, s) {
			sc.InsertStatSlotByOrder(s)
		}
	}
}

// RegisterGlobalStatPrepareSlot is not thread safe, and user must call RegisterGlobalStatPrepareSlot when initializing sentinel running environment
func RegisterGlobalStatPrepareSlot(slot base.StatPrepareSlot) {
	for _, s := range globalStatPrepareSlots {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalStatPrepareSlots = append(globalStatPrepareSlots, slot)
}

// RegisterGlobalRuleCheckSlot is not thread safe, and user must call RegisterGlobalRuleCheckSlot when initializing sentinel running environment
func RegisterGlobalRuleCheckSlot(slot base.RuleCheckSlot) {
	for _, s := range globalRuleCheckSlots {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalRuleCheckSlots = append(globalRuleCheckSlots, slot)
}

// RegisterGlobalStatSlot is not thread safe, and user must call RegisterGlobalStatSlot when initializing sentinel running environment
func RegisterGlobalStatSlot(slot base.StatSlot) {
	for _, s := range globalStatSlot {
		if s.Name() == slot.Name() {
			return
		}
	}
	globalStatSlot = append(globalStatSlot, slot)
}

func newResourceSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.InsertStatPrepareSlotByOrder(stat.DefaultResourceNodePrepareSlot)

	sc.InsertStatSlotByOrder(stat.DefaultSlot)
	sc.InsertStatSlotByOrder(log.DefaultSlot)
	registerCustomGlobalSlotsToSc(sc)
	return sc
}

func RegisterStatPrepareSlotForResource(rsName string, slot base.StatPrepareSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if base.ValidateStatPrepareSlotNaming(sc, slot) {
		sc.InsertStatPrepareSlotByOrder(slot)
	}
}

func RegisterRuleCheckSlotForResource(rsName string, slot base.RuleCheckSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if base.ValidateRuleCheckSlotNaming(sc, slot) {
		sc.InsertRuleCheckSlotByOrder(slot)
	}
}

func RegisterStatSlotForResource(rsName string, slot base.StatSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if base.ValidateStatSlotNaming(sc, slot) {
		sc.InsertStatSlotByOrder(slot)
	}
}

func GetResourceSlotChain(rsName string) *base.SlotChain {
	rsSlotChainLock.RLock()
	defer rsSlotChainLock.RUnlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		return nil
	}

	return sc
}
