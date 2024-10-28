package micro

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/server"

	"github.com/alibaba/sentinel-golang/core/base"
)

type (
	Option func(*options)

	options struct {
		clientResourceExtract func(context.Context, client.Request) string
		serverResourceExtract func(context.Context, server.Request) string

		streamClientResourceExtract func(context.Context, client.Request) string
		streamServerResourceExtract func(server.Stream) string

		clientBlockFallback func(context.Context, client.Request, *base.BlockError) error
		serverBlockFallback func(context.Context, server.Request, *base.BlockError) error

		streamClientBlockFallback func(context.Context, client.Request, *base.BlockError) (client.Stream, error)
		streamServerBlockFallback func(server.Stream, *base.BlockError) server.Stream

		enableOutlier func(ctx context.Context) bool
	}
)

// WithClientResourceExtractor sets the resource extractor of unary client request.
// The second string parameter is the full method name of current invocation.
func WithClientResourceExtractor(fn func(context.Context, client.Request) string) Option {
	return func(opts *options) {
		opts.clientResourceExtract = fn
	}
}

// WithServerResourceExtractor sets the resource extractor of unary server request.
func WithServerResourceExtractor(fn func(context.Context, server.Request) string) Option {
	return func(opts *options) {
		opts.serverResourceExtract = fn
	}
}

// WithStreamClientResourceExtractor sets the resource extractor of stream client request.
func WithStreamClientResourceExtractor(fn func(context.Context, client.Request) string) Option {
	return func(opts *options) {
		opts.streamClientResourceExtract = fn
	}
}

// WithStreamServerResourceExtractor sets the resource extractor of stream server request.
func WithStreamServerResourceExtractor(fn func(server.Stream) string) Option {
	return func(opts *options) {
		opts.streamServerResourceExtract = fn
	}
}

// WithClientBlockFallback sets the block fallback handler of unary client request.
// The second string parameter is the full method name of current invocation.
func WithClientBlockFallback(fn func(context.Context, client.Request, *base.BlockError) error) Option {
	return func(opts *options) {
		opts.clientBlockFallback = fn
	}
}

// WithServerBlockFallback sets the block fallback handler of unary server request.
func WithServerBlockFallback(fn func(context.Context, server.Request, *base.BlockError) error) Option {
	return func(opts *options) {
		opts.serverBlockFallback = fn
	}
}

// WithStreamClientBlockFallback sets the block fallback handler of stream client request.
func WithStreamClientBlockFallback(fn func(context.Context, client.Request, *base.BlockError) (client.Stream, error)) Option {
	return func(opts *options) {
		opts.streamClientBlockFallback = fn
	}
}

// WithStreamServerBlockFallback sets the block fallback handler of stream server request.
func WithStreamServerBlockFallback(fn func(server.Stream, *base.BlockError) server.Stream) Option {
	return func(opts *options) {
		opts.streamServerBlockFallback = fn
	}
}

// WithEnableOutlier sets whether to enable outlier ejection
func WithEnableOutlier(fn func(ctx context.Context) bool) Option {
	return func(opts *options) {
		opts.enableOutlier = fn
	}
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
