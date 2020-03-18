package rocketmq

import (
	"context"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type (
	Fallback func(ctx context.Context, req, reply interface{}, next primitive.Invoker, blockError *base.BlockError) error
	Option func(*options)
	options struct {
		consumerResourceExtract func(*primitive.ConsumeMessageContext) string
		providerResourceExtract func(*primitive.ProducerCtx) string
		blockFallback Fallback
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
func WithBlockFallback(fn Fallback) Option {
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
