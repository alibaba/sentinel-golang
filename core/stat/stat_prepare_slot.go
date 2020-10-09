package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

type ResourceNodePrepareSlot struct {
}

func (s *ResourceNodePrepareSlot) Prepare(ctx *base.EntryContext) {
	node := GetOrCreateResourceNode(ctx.Resource.Name(), ctx.Resource.Classification())
	// Set the resource node to the context.
	ctx.StatNode = node
}
