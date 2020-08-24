package api

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
)

var entryOptsPool = sync.Pool{
	New: func() interface{} {
		return &EntryOptions{
			resourceType: base.ResTypeCommon,
			entryType:    base.Outbound,
			acquireCount: 1,
			flag:         0,
			slotChain:    nil,
			args:         nil,
			attachments:  nil,
		}
	},
}

// EntryOptions represents the options of a Sentinel resource entry.
type EntryOptions struct {
	resourceType base.ResourceType
	entryType    base.TrafficType
	acquireCount uint32
	flag         int32
	slotChain    *base.SlotChain
	args         []interface{}
	attachments  map[interface{}]interface{}
}

func (o *EntryOptions) Reset() {
	o.resourceType = base.ResTypeCommon
	o.entryType = base.Outbound
	o.acquireCount = 1
	o.flag = 0
	o.slotChain = nil
	o.args = nil
	o.attachments = nil
}

type EntryOption func(*EntryOptions)

// WithResourceType sets the resource entry with the given resource type.
func WithResourceType(resourceType base.ResourceType) EntryOption {
	return func(opts *EntryOptions) {
		opts.resourceType = resourceType
	}
}

// WithTrafficType sets the resource entry with the given traffic type.
func WithTrafficType(entryType base.TrafficType) EntryOption {
	return func(opts *EntryOptions) {
		opts.entryType = entryType
	}
}

// WithAcquireCount sets the resource entry with the given batch count (by default 1).
func WithAcquireCount(acquireCount uint32) EntryOption {
	return func(opts *EntryOptions) {
		opts.acquireCount = acquireCount
	}
}

// WithFlag sets the resource entry with the given additional flag.
func WithFlag(flag int32) EntryOption {
	return func(opts *EntryOptions) {
		opts.flag = flag
	}
}

// WithArgs sets the resource entry with the given additional parameters.
func WithArgs(args ...interface{}) EntryOption {
	return func(opts *EntryOptions) {
		opts.args = append(opts.args, args...)
	}
}

// WithSlotChain sets the slot chain.
func WithSlotChain(chain *base.SlotChain) EntryOption {
	return func(opts *EntryOptions) {
		opts.slotChain = chain
	}
}

// WithAttachment set the resource entry with the given k-v pair
func WithAttachment(key interface{}, value interface{}) EntryOption {
	return func(opts *EntryOptions) {
		if opts.attachments == nil {
			opts.attachments = make(map[interface{}]interface{})
		}
		opts.attachments[key] = value
	}
}

// WithAttachment set the resource entry with the given k-v pairs
func WithAttachments(data map[interface{}]interface{}) EntryOption {
	return func(opts *EntryOptions) {
		if opts.attachments == nil {
			opts.attachments = make(map[interface{}]interface{})
		}
		for key, value := range data {
			opts.attachments[key] = value
		}
	}
}

// Entry is the basic API of Sentinel.
func Entry(resource string, opts ...EntryOption) (*base.SentinelEntry, *base.BlockError) {
	options := entryOptsPool.Get().(*EntryOptions)
	options.slotChain = globalSlotChain

	for _, opt := range opts {
		opt(options)
	}

	return entry(resource, options)
}

func entry(resource string, options *EntryOptions) (*base.SentinelEntry, *base.BlockError) {
	rw := base.NewResourceWrapper(resource, options.resourceType, options.entryType)
	sc := options.slotChain

	if sc == nil {
		return base.NewSentinelEntry(nil, rw, nil), nil
	}
	// Get context from pool.
	ctx := sc.GetPooledContext()
	ctx.Resource = rw
	ctx.Input.AcquireCount = options.acquireCount
	ctx.Input.Flag = options.flag
	if len(options.args) != 0 {
		ctx.Input.Args = options.args
	}
	if len(options.attachments) != 0 {
		ctx.Input.Attachments = options.attachments
	}
	options.Reset()
	entryOptsPool.Put(options)
	e := base.NewSentinelEntry(ctx, rw, sc)
	ctx.SetEntry(e)
	r := sc.Entry(ctx)
	if r == nil {
		// This indicates internal error in some slots, so just pass
		return e, nil
	}
	if r.Status() == base.ResultStatusBlocked {
		// r will be put to Pool in calling Exit()
		// must finish the lifecycle of r.
		blockErr := base.NewBlockErrorFromDeepCopy(r.BlockError())
		e.Exit()
		return nil, blockErr
	}

	return e, nil
}
