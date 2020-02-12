package api

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

// Options represents the options of a Sentinel resource entry.
type Options struct {
	resourceType base.ResourceType
	entryType    base.TrafficType
	acquireCount uint32
	flag         int32
	slotChain    *base.SlotChain
	args         []interface{}
}

type Option func(*Options)

// WithResourceType sets the resource entry with the given resource type.
func WithResourceType(resourceType base.ResourceType) Option {
	return func(opts *Options) {
		opts.resourceType = resourceType
	}
}

// WithTrafficType sets the resource entry with the given traffic type.
func WithTrafficType(entryType base.TrafficType) Option {
	return func(opts *Options) {
		opts.entryType = entryType
	}
}

// WithAcquireCount sets the resource entry with the given batch count (by default 1).
func WithAcquireCount(acquireCount uint32) Option {
	return func(opts *Options) {
		opts.acquireCount = acquireCount
	}
}

// WithFlag sets the resource entry with the given additional flag.
func WithFlag(flag int32) Option {
	return func(opts *Options) {
		opts.flag = flag
	}
}

// WithArgs sets the resource entry with the given additional parameters.
func WithArgs(args ...interface{}) Option {
	return func(opts *Options) {
		opts.args = append(opts.args, args...)
	}
}
