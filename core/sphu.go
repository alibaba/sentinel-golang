package core

func Entry(name string, opts ...EntryOption) *CtxEntry {
	var options = defaultEntryOptions()
	for _, opt := range opts {
		opt(&options)
	}
	rw := &ResourceWrapper{
		ResourceName: name,
		FlowType:     options.trafficType,
	}

	sc := GetDefaultSlotChain()
	// get context
	ctx := sc.GetContext()
	ctx.ResWrapper = rw
	ctx.Count = options.count
	ctx.Entry = NewCtxEntry(ctx, rw, sc, ctx.StatNode)

	sc.entry(ctx)

	return ctx.Entry
}

type (
	EntryOptions struct {
		count uint64
		trafficType TrafficType
	}
	EntryOption func(options *EntryOptions)
)

func defaultEntryOptions() EntryOptions {
	return EntryOptions{
		count:       1,
		trafficType: InBound,
	}
}

func WithCount(count uint64) EntryOption {
	return func(options *EntryOptions) {
		options.count = count
	}
}

func WithTrafficType(trafficType TrafficType) EntryOption {
	return func(options *EntryOptions) {
		options.trafficType = trafficType
	}
}
