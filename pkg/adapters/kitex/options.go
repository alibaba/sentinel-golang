package kitex

import (
	"context"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

type Option struct {
	F func(o *options)
}

type options struct {
	ResourceExtract func(ctx context.Context, req, resp interface{}) string
	BlockFallback   func(ctx context.Context, req, resp interface{}, blockErr error) error
}

func DefaultBlockFallback(ctx context.Context, req, resp interface{}, blockErr error) error {
	return blockErr
}

func DefaultResourceExtract(ctx context.Context, req, resp interface{}) string {
	ri := rpcinfo.GetRPCInfo(ctx)
	return ri.To().ServiceName() + ":" + ri.To().Method()
}

func newOptions(opts []Option) *options {
	o := &options{
		ResourceExtract: DefaultResourceExtract,
		BlockFallback:   DefaultBlockFallback,
	}
	o.Apply(opts)
	return o
}

func (o *options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

// WithResourceExtract sets the resource extractor
func WithResourceExtract(f func(ctx context.Context, req, resp interface{}) string) Option {
	return Option{F: func(o *options) {
		o.ResourceExtract = f
	}}
}

// WithBlockFallback sets the fallback handler
func WithBlockFallback(f func(ctx context.Context, req, resp interface{}, blockErr error) error) Option {
	return Option{func(o *options) {
		o.BlockFallback = f
	}}
}
