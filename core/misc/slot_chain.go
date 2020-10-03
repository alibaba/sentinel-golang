package misc

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/log"
	"github.com/alibaba/sentinel-golang/core/stat"
)

var (
	rsSlotChainLock sync.RWMutex
	rsSlotChain = make(map[string]*base.SlotChain, 8)
)

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

func newSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(stat.DefaultResourceNodePrepareSlot)

	sc.AddStatSlotLast(stat.DefaultSlot)
	sc.AddStatSlotLast(log.DefaultSlot)

	return sc
}

func RegisterResourceRuleCheckSlot(rsName string, slot base.RuleCheckSlot) {
	rsSlotChainLock.Lock()
	defer rsSlotChainLock.Unlock()

	sc, ok := rsSlotChain[rsName]
	if !ok {
		sc = newSlotChain()
	}

	if !validateRuleCheckSlot(sc, slot) {
		sc.AddRuleCheckSlotLast(slot)
	}
	if !ok {
		rsSlotChain[rsName] = sc
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
		sc = newSlotChain()
	}

	if !validateStatSlot(sc, slot) {
		sc.AddStatSlotLast(slot)
	}
	if !ok {
		rsSlotChain[rsName] = sc
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
