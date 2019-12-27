package api

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

// defaultSlotChain is a default slot chain built by framework
// defaultSlotChain is global unique chain
var defaultSlotChain = buildDefaultSlotChain()

func buildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	// insert slots
	return sc
}

func DefaultSlotChain() *base.SlotChain {
	return defaultSlotChain
}
