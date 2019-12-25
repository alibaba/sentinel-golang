package api

import "github.com/sentinel-group/sentinel-golang/core"

// defaultSlotChain is a default slot chain built by framework
// defaultSlotChain is global unique chain
var defaultSlotChain = buildDefaultSlotChain()

func buildDefaultSlotChain() *core.SlotChain {
	sc := core.NewSlotChain()
	// insert slots
	return sc
}

func DefaultSlotChain() *core.SlotChain {
	return defaultSlotChain
}
