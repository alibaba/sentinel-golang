package rocketmq

import (
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type (
	Option func(*options)
	options struct {
		resourceExtract func(*primitive.ProducerCtx) string
		blockFallback primitive.Interceptor
	}
)

func WithResourceExtract(fn func(ctx *primitive.ProducerCtx) string) Option {
	return func(options *options) {
		options.resourceExtract = fn
	}
}

func WithBlockFallback(fn primitive.Interceptor) Option {
	return func(options *options) {
		options.blockFallback = fn
	}
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}
	return optCopy
}
