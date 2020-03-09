package rocketmq

import (
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type (
	Option func(*options)
	options struct {
		consumerResourceExtract func(*primitive.ConsumeMessageContext) string
		providerResourceExtract func(*primitive.ProducerCtx) string
		blockFallback primitive.Interceptor
	}
)

// WithConsumerResourceExtract set consumerResourceExtract
func WithConsumerResourceExtract(fn func(ctx *primitive.ConsumeMessageContext) string) Option {
	return func(options *options) {
		options.consumerResourceExtract = fn
	}
}

// WithProviderResourceExtract set providerResourceExtract
func WithProviderResourceExtract(fn func(ctx *primitive.ProducerCtx) string) Option {
	return func(options *options) {
		options.providerResourceExtract = fn
	}
}

// WithBlockFallback set blockFallback
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
