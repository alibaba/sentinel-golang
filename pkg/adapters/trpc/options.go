package trpc

import (
	"context"

	"github.com/alibaba/sentinel-golang/core/base"
	"trpc.group/trpc-go/trpc-go/codec"
)

// Option is a function that configures the options.
type Option struct {
	F func(o *options)
}

type options struct {
	resourceExtract func(ctx context.Context, req interface{}) string
	blockFallback   func(ctx context.Context, req interface{}, blockErr *base.BlockError) (interface{}, error)
}

// DefaultResourceExtract extracts resource name from trpc context.
// Default resource name format is: {CalleeServiceName}:{ServerRPCName}
// For server side, it uses CalleeServiceName and ServerRPCName.
// For client side, it uses CalleeServiceName and ClientRPCName.
func DefaultResourceExtract(ctx context.Context, req interface{}) string {
	msg := codec.Message(ctx)
	serviceName := msg.CalleeServiceName()
	rpcName := msg.ServerRPCName()
	if rpcName == "" {
		rpcName = msg.ClientRPCName()
	}
	if serviceName == "" {
		return rpcName
	}
	if rpcName == "" {
		return serviceName
	}
	return serviceName + ":" + rpcName
}

// DefaultBlockFallback is the default block fallback function.
// It returns nil response and the block error.
func DefaultBlockFallback(ctx context.Context, req interface{}, blockErr *base.BlockError) (interface{}, error) {
	return nil, blockErr
}

func newOptions(opts []Option) *options {
	o := &options{
		resourceExtract: DefaultResourceExtract,
		blockFallback:   DefaultBlockFallback,
	}
	o.Apply(opts)
	return o
}

// Apply applies the given options.
func (o *options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

// WithResourceExtract sets the resource extractor function.
// The function extracts resource name from context and request.
func WithResourceExtract(f func(ctx context.Context, req interface{}) string) Option {
	return Option{F: func(o *options) {
		o.resourceExtract = f
	}}
}

// WithBlockFallback sets the block fallback handler function.
// The function is called when the request is blocked by Sentinel.
func WithBlockFallback(f func(ctx context.Context, req interface{}, blockErr *base.BlockError) (interface{}, error)) Option {
	return Option{F: func(o *options) {
		o.blockFallback = f
	}}
}
