package api

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

func Entry(resource string, opts ...Option) (*base.SentinelEntry, *base.BlockError) {
	var options = Options{
		resourceType: base.ResTypeCommon,
		entryType:    base.Outbound,
		acquireCount: 1,
		flag:         0,
		slotChain: 	defaultSlotChain,
		args:         []interface{}{},
	}
	for _, opt := range opts {
		opt(&options)
	}

	return entry(resource, &options)
}

func entry(resource string, options *Options) (*base.SentinelEntry, *base.BlockError) {
	rw := base.NewResourceWrapper(resource, options.resourceType, options.entryType)
	sc := options.slotChain

	if sc == nil {
		return base.NewSentinelEntry(nil, rw, nil), nil
	}
	// Get context from pool.
	ctx := sc.GetPooledContext()
	ctx.Resource = rw
	ctx.Input = &base.SentinelInput{
		AcquireCount: options.acquireCount,
		Flag:         options.flag,
		Args:         options.args,
	}

	e := base.NewSentinelEntry(ctx, rw, sc)

	r := sc.Entry(ctx)
	if r == nil {
		// This indicates internal error in some slots, so just pass
		return e, nil
	}
	if r.Status() == base.ResultStatusBlocked {
		e.Exit()
		return nil, r.BlockError()
	}

	return e, nil
}
