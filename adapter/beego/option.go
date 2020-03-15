package beego

import (
	"github.com/astaxie/beego/context"
)

type (
	Option func(*options)
	options struct {
		resourceExtract func(*context.Context) string
		blockFallback func(*context.Context)
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}

	return optCopy
}

// WithResourceExtractor set resourceExtractor
func WithResourceExtractor(fn func(*context.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback set blockFallback
func WithBlockFallback(fn func(ctx *context.Context)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
