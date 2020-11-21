package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

const (
	PrepareSlotName  = "sentinel-core-stat-resource-node-prepare-slot"
	PrepareSlotOrder = 1000
)

var (
	DefaultResourceNodePrepareSlot = &ResourceNodePrepareSlot{}
)

type ResourceNodePrepareSlot struct {
}

func (s *ResourceNodePrepareSlot) Name() string {
	return PrepareSlotName
}

func (s *ResourceNodePrepareSlot) Order() uint32 {
	return PrepareSlotOrder
}

func (s *ResourceNodePrepareSlot) Prepare(ctx *base.EntryContext) {
	node := GetOrCreateResourceNode(ctx.Resource.Name(), ctx.Resource.Classification())
	// Set the resource node to the context.
	ctx.StatNode = node
}
