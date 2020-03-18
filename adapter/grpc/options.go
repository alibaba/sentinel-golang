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

// WithUnaryClientResourceExtractor set unaryClientResourceExtract
func WithUnaryClientResourceExtractor(fn func(context.Context, string, interface{}, *grpc.ClientConn) string) Option {
	return func(opts *options) {
		opts.unaryClientResourceExtract = fn
	}
}

// WithUnaryServerResourceExtractor set unaryServerResourceExtract
func WithUnaryServerResourceExtractor(fn func(context.Context, interface{}, *grpc.UnaryServerInfo) string) Option {
	return func(opts *options) {
		opts.unaryServerResourceExtract = fn
	}
}

// WithStreamClientResourceExtractor set streamClientResourceExtract
func WithStreamClientResourceExtractor(fn func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string) string) Option {
	return func(opts *options) {
		opts.streamClientResourceExtract = fn
	}
}

// WithStreamServerResourceExtractor set streamServerResourceExtract
func WithStreamServerResourceExtractor(fn func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo) string) Option {
	return func(opts *options) {
		opts.streamServerResourceExtract = fn
	}
}

// WithUnaryClientBlockFallback set unaryClientBlockFallback
func WithUnaryClientBlockFallback(fn func(context.Context, string, interface{}, *grpc.ClientConn, *base.BlockError) error) Option {
	return func(opts *options) {
		opts.unaryClientBlockFallback = fn
	}
}

// WithUnaryServerBlockFallback set unaryServerBlockFallback
func WithUnaryServerBlockFallback(fn func(context.Context, interface{}, *grpc.UnaryServerInfo, *base.BlockError) (interface{}, error)) Option {
	return func(opts *options) {
		opts.unaryServerBlockFallback = fn
	}
}

// WithStreamClientBlockFallback set streamClientBlockFallback
func WithStreamClientBlockFallback(fn func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, *base.BlockError) (grpc.ClientStream, error)) Option {
	return func(opts *options) {
		opts.streamClientBlockFallback = fn
	}
}

// WithStreamServerBlockFallback set streamServerBlockFallback
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
