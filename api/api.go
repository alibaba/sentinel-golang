package api

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

func Entry(name string) (*base.SentinelEntry, *base.BlockError) {
	return EntryWithType(name, base.ResTypeCommon, base.Outbound)
}

func EntryWithType(resource string, resType base.ResourceType, entryType base.TrafficType) (*base.SentinelEntry, *base.BlockError) {
	return EntryWithTypeAndCount(resource, resType, entryType, 1)
}

func EntryWithTypeAndCount(resource string, resType base.ResourceType, entryType base.TrafficType, acquireCount uint32) (*base.SentinelEntry, *base.BlockError) {
	return EntryWithArgs(resource, resType, entryType, acquireCount, 0)
}

func EntryWithArgs(resource string, resType base.ResourceType, entryType base.TrafficType, acquireCount uint32, flag int32, args ...interface{}) (*base.SentinelEntry, *base.BlockError) {
	return entryWithArgsAndChain(resource, resType, entryType, acquireCount, flag, DefaultSlotChain(), args)
}

func entryWithArgsAndChain(resource string, resType base.ResourceType, entryType base.TrafficType, acquireCount uint32, flag int32, sc *base.SlotChain, args ...interface{}) (*base.SentinelEntry, *base.BlockError) {
	rw := base.NewResourceWrapper(resource, resType, entryType)

	if sc == nil {
		return base.NewSentinelEntry(nil, rw, nil), nil
	}
	// Get context from pool.
	ctx := sc.GetPooledContext()
	ctx.Resource = rw
	ctx.Input = &base.SentinelInput{
		AcquireCount: acquireCount,
		Flag:         flag,
		Args:         args,
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
