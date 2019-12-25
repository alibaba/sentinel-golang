package api

import "github.com/sentinel-group/sentinel-golang/core"

func Entry(name string) (*core.SentinelEntry, *core.BlockError) {
	return EntryWithType(name, core.ResTypeCommon, core.Outbound)
}

func EntryWithType(resource string, resType core.ResourceType, entryType core.TrafficType) (*core.SentinelEntry, *core.BlockError) {
	return EntryWithTypeAndCount(resource, resType, entryType, 1)
}

func EntryWithTypeAndCount(resource string, resType core.ResourceType, entryType core.TrafficType, acquireCount uint32) (*core.SentinelEntry, *core.BlockError) {
	return EntryWithArgs(resource, resType, entryType, acquireCount, 0)
}

func EntryWithArgs(resource string, resType core.ResourceType, entryType core.TrafficType, acquireCount uint32, flag int32, args ...interface{}) (*core.SentinelEntry, *core.BlockError) {
	return entryWithArgsAndChain(resource, resType, entryType, acquireCount, flag, DefaultSlotChain(), args)
}

func entryWithArgsAndChain(resource string, resType core.ResourceType, entryType core.TrafficType, acquireCount uint32, flag int32, sc *core.SlotChain, args ...interface{}) (*core.SentinelEntry, *core.BlockError) {
	rw := core.NewResourceWrapper(resource, resType, entryType)

	if sc == nil {
		return core.NewSentinelEntry(nil, rw, nil), nil
	}
	// Get context from pool.
	ctx := sc.GetPooledContext()
	ctx.Resource = rw
	ctx.Input = &core.SentinelInput{
		AcquireCount: acquireCount,
		Flag:         flag,
		Args:         args,
	}

	e := core.NewSentinelEntry(ctx, rw, sc)

	r := sc.Entry(ctx)
	if r == nil {
		// This indicates internal error in some slots, so just pass
		return e, nil
	}
	if r.Status() == core.ResultStatusBlocked {
		e.Exit()
		return nil, r.BlockError()
	}

	return e, nil
}
