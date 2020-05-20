package base

import (
	"sync"
)

type SentinelEntry struct {
	res *ResourceWrapper
	// one entry with one context
	ctx *EntryContext
	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain

	exitCtl sync.Once
}

func NewSentinelEntry(ctx *EntryContext, rw *ResourceWrapper, sc *SlotChain) *SentinelEntry {
	return &SentinelEntry{
		res: rw,
		ctx: ctx,
		sc:  sc,
	}
}

func (e *SentinelEntry) SetError(err error) {
	if e.ctx != nil {
		e.ctx.SetError(err)
	}
}

func (e *SentinelEntry) Context() *EntryContext {
	return e.ctx
}

func (e *SentinelEntry) Resource() *ResourceWrapper {
	return e.res
}

type ExitOptions struct {
	err error
}
type ExitOption func(*ExitOptions)

func WithError(err error) ExitOption {
	return func(opts *ExitOptions) {
		opts.err = err
	}
}

func (e *SentinelEntry) Exit(exitOps ...ExitOption) {
	var options = ExitOptions{
		err: nil,
	}
	for _, opt := range exitOps {
		opt(&options)
	}
	ctx := e.ctx
	if options.err != nil {
		ctx.SetError(options.err)
	}
	e.exitCtl.Do(func() {
		if e.sc != nil {
			e.sc.exit(ctx)
			e.sc.RefurbishContext(ctx)
		}
	})
}
