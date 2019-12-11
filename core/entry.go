package core

import "github.com/sentinel-group/sentinel-golang/util"

// CtxEntry means EntryContext entry,
type CtxEntry struct {
	createTime uint64
	rs         *ResourceWrapper
	// one entry with one context
	ctx *EntryContext
	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain
	// caller node
	originNode Node
	// current resource node
	currentNode Node
}

func NewCtxEntry(ctx *EntryContext, rw *ResourceWrapper, sc *SlotChain, cn Node) *CtxEntry {
	return &CtxEntry{
		createTime:  util.CurrentTimeMillis(),
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

func (e *CtxEntry) exitForContext(ctx *EntryContext, count int32) {
	if e.sc != nil {
		e.sc.exit(ctx)
	}
}

func (e *CtxEntry) GetCurrentNode() Node {
	return e.currentNode
}
