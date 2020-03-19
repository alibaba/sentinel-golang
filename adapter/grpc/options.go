package grpc

import (
	"context"

	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

type (
	Option func(*options)

	options struct {
		unaryClientResourceExtract func(context.Context, string, interface{}, *grpc.ClientConn) string
		unaryServerResourceExtract func(context.Context, interface{}, *grpc.UnaryServerInfo) string

		streamClientResourceExtract func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string) string
		streamServerResourceExtract func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo) string

		unaryClientBlockFallback func(context.Context, string, interface{}, *grpc.ClientConn, *base.BlockError) error
		unaryServerBlockFallback func(context.Context, interface{}, *grpc.UnaryServerInfo, *base.BlockError) (interface{}, error)

		streamClientBlockFallback func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, *base.BlockError) (grpc.ClientStream, error)
		streamServerBlockFallback func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, *base.BlockError) error
	}
)

// WithUnaryClientResourceExtractor sets the resource extractor of unary client request.
// The second string parameter is the full method name of current invocation.
func WithUnaryClientResourceExtractor(fn func(context.Context, string, interface{}, *grpc.ClientConn) string) Option {
	return func(opts *options) {
		opts.unaryClientResourceExtract = fn
	}
}

// WithUnaryServerResourceExtractor sets the resource extractor of unary server request.
func WithUnaryServerResourceExtractor(fn func(context.Context, interface{}, *grpc.UnaryServerInfo) string) Option {
	return func(opts *options) {
		opts.unaryServerResourceExtract = fn
	}
}

// WithStreamClientResourceExtractor sets the resource extractor of stream client request.
func WithStreamClientResourceExtractor(fn func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string) string) Option {
	return func(opts *options) {
		opts.streamClientResourceExtract = fn
	}
}

// WithStreamServerResourceExtractor sets the resource extractor of stream server request.
func WithStreamServerResourceExtractor(fn func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo) string) Option {
	return func(opts *options) {
		opts.streamServerResourceExtract = fn
	}
}

// WithUnaryClientBlockFallback sets the block fallback handler of unary client request.
// The second string parameter is the full method name of current invocation.
func WithUnaryClientBlockFallback(fn func(context.Context, string, interface{}, *grpc.ClientConn, *base.BlockError) error) Option {
	return func(opts *options) {
		opts.unaryClientBlockFallback = fn
	}
}

// WithUnaryServerBlockFallback sets the block fallback handler of unary server request.
func WithUnaryServerBlockFallback(fn func(context.Context, interface{}, *grpc.UnaryServerInfo, *base.BlockError) (interface{}, error)) Option {
	return func(opts *options) {
		opts.unaryServerBlockFallback = fn
	}
}

// WithStreamClientBlockFallback sets the block fallback handler of stream client request.
func WithStreamClientBlockFallback(fn func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, *base.BlockError) (grpc.ClientStream, error)) Option {
	return func(opts *options) {
		opts.streamClientBlockFallback = fn
	}
}

// WithStreamServerBlockFallback sets the block fallback handler of stream server request.
func WithStreamServerBlockFallback(fn func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, *base.BlockError) error) Option {
	return func(opts *options) {
		opts.streamServerBlockFallback = fn
	}
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
