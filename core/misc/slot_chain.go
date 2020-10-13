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

func GlobalStatPrepareSlots() []base.StatPrepareSlot {
	return globalStatPrepareSlots
}

func GlobalRuleCheckSlots() []base.RuleCheckSlot {
	return globalRuleCheckSlots
}

func GlobalStatSlot() []base.StatSlot {
	return globalStatSlot
}

func RegisterCustomGlobalSlotsToSc(sc *base.SlotChain) {
	if sc == nil {
		return
	}
	for _, s := range globalStatPrepareSlots {
		if !validateStatPrepareSlot(sc, s) {
			sc.AddStatPrepareSlotLast(s)
		}
	}
	for _, s := range globalRuleCheckSlots {
		if !validateRuleCheckSlot(sc, s) {
			sc.AddRuleCheckSlotLast(s)
		}
	}
	for _, s := range globalStatSlot {
		if !validateStatSlot(sc, s) {
			sc.AddStatSlotLast(s)
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
	sc.AddStatPrepareSlotLast(stat.DefaultResourceNodePrepareSlot)

	sc.AddStatSlotLast(stat.DefaultSlot)
	sc.AddStatSlotLast(log.DefaultSlot)
	RegisterCustomGlobalSlotsToSc(sc)
	return sc
}

func validateStatPrepareSlot(sc *base.SlotChain, s base.StatPrepareSlot) bool {
	flag := false
	f := func(slot base.StatPrepareSlot) {
		if slot.Name() == s.Name() {
			flag = true
		}
	}
	sc.RangeStatPrepareSlot(f)

	return flag
}

func RegisterResourceStatPrepareSlot(rsName string, slot base.StatPrepareSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if !validateStatPrepareSlot(sc, slot) {
		sc.AddStatPrepareSlotLast(slot)
	}
}

func validateRuleCheckSlot(sc *base.SlotChain, s base.RuleCheckSlot) bool {
	flag := false
	f := func(slot base.RuleCheckSlot) {
		if slot.Name() == s.Name() {
			flag = true
		}
	}
	sc.RangeRuleCheckSlot(f)

	return flag
}

func RegisterResourceRuleCheckSlot(rsName string, slot base.RuleCheckSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if !validateRuleCheckSlot(sc, slot) {
		sc.AddRuleCheckSlotLast(slot)
	}
}

func validateStatSlot(sc *base.SlotChain, s base.StatSlot) bool {
	flag := false
	f := func(slot base.StatSlot) {
		if slot.Name() == s.Name() {
			flag = true
		}
	}
	sc.RangeStatSlot(f)

	return flag
}

func RegisterResourceStatSlot(rsName string, slot base.StatSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newResourceSlotChain()
		rsSlotChain[rsName] = sc
	}

	if !validateStatSlot(sc, slot) {
		sc.AddStatSlotLast(slot)
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
