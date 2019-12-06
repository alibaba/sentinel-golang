package core

// Name based Entry
func Entry(name string) *CtxEntry {
	return Entry2(name, 1)
}

func Entry2(name string, count uint64) *CtxEntry {
	return Entry3(name, count, InBound)
}

func Entry3(name string, count uint64, entryType TrafficType) *CtxEntry {
	rw := &ResourceWrapper{
		ResourceName: name,
		FlowType:     entryType,
	}

	sc := GetDefaultSlotChain()
	// get context
	ctx := sc.GetContext()
	ctx.ResWrapper = rw
	ctx.Count = count
	ctx.Entry = NewCtEntry(ctx, rw, sc, ctx.StatNode)

	sc.entry(ctx)

	return ctx.Entry
}
