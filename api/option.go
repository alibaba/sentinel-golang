package api

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

type Options struct {
	resourceType base.ResourceType
	entryType base.TrafficType
	acquireCount uint32
	flag int32
	slotChain *base.SlotChain
	args []interface{}
}

type Option func(*Options)

func WithResourceType(resourceType base.ResourceType) Option {
	return func(opts *Options) {
		opts.resourceType = resourceType
	}
}

func WithEntryType(entryType base.TrafficType) Option {
	return func(opts *Options) {
		opts.entryType = entryType
	}
}

func WithAcquireCount(acquireCount uint32) Option {
	return func(opts *Options) {
		opts.acquireCount = acquireCount
	}
}

func WithFlag(flag int32) Option {
	return func(opts *Options) {
		opts.flag = flag
	}
}

func WithArgs(args ...interface{}) Option {
	return func(opts *Options) {
		opts.args = append(opts.args, args...)
	}
}
