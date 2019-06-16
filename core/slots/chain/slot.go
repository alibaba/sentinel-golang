package chain

import (
	"context"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type Slot interface {
	IsContinue(lastResult base.SlotResult, ctx context.Context) bool

	Entry(ctx context.Context, resourceWrap *base.ResourceWrapper, node *base.DefaultNode, count uint32) base.SlotResult

	Exit(ctx context.Context, resourceWrap *base.ResourceWrapper, count uint32)
}
