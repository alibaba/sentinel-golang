package core

// TrafficType describe the traffic type: Inbound or OutBound
type TrafficType int32

const (
	InBound TrafficType = iota
	OutBound
)

type ResourceWrapper struct {
	// global unique resource name
	ResourceName string
	// InBound or OutBound
	FlowType TrafficType
}

// CtxEntry means Context entry,
type CtxEntry struct {
	createTime uint64
	rs         *ResourceWrapper
	// one entry with one context
	ctx *Context
	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain
	// caller node
	originNode node
	// current resource node
	currentNode node
}

func NewCtEntry(ctx *Context, rw *ResourceWrapper, sc *SlotChain, cn node) *CtxEntry {
	return &CtxEntry{
		createTime:  GetTimeMilli(),
		rs:          rw,
		ctx:         ctx,
		sc:          sc,
		currentNode: cn,
	}
}

func (e *CtxEntry) Exit() {
	e.ExitWithCnt(1)
}

func (e *CtxEntry) ExitWithCnt(count int32) {
	e.exitForContext(e.ctx, count)
}

func (e *CtxEntry) exitForContext(ctx *Context, count int32) {
	if e.sc != nil {
		e.sc.exit(ctx)
	}
}

func (e *CtxEntry) GetCurrentNode() node {
	return e.currentNode
}
