package kratos

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
)

type Option struct {
	F func(o *options)
}

type options struct {
	ResourceExtract func(ctx context.Context, req interface{}) string
	BlockFallback   func(ctx context.Context, req interface{}, blockErr error) (interface{}, error)
	EnableOutlier   func(ctx context.Context) bool
}

func DefaultResourceExtract(ctx context.Context, req interface{}) string {
	if v, ok := transport.FromClientContext(ctx); ok {
		return v.Operation()
	}
	panic("operation is empty")
}

func DefaultBlockFallback(ctx context.Context, req interface{}, blockErr error) (interface{}, error) {
	return nil, blockErr
}

func DefaultEnableOutlier(ctx context.Context) bool {
	return false
}

func newOptions(opts []Option) *options {
	o := &options{
		ResourceExtract: DefaultResourceExtract,
		BlockFallback:   DefaultBlockFallback,
		EnableOutlier:   DefaultEnableOutlier,
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
func WithResourceExtract(f func(ctx context.Context, req interface{}) string) Option {
	return Option{F: func(o *options) {
		o.ResourceExtract = f
	}}
}

// WithBlockFallback sets the fallback handler
func WithBlockFallback(f func(ctx context.Context, req interface{}, blockErr error) (interface{}, error)) Option {
	return Option{func(o *options) {
		o.BlockFallback = f
	}}
}

// WithEnableOutlier sets whether to enable outlier ejection
func WithEnableOutlier(f func(ctx context.Context) bool) Option {
	return Option{func(o *options) {
		o.EnableOutlier = f
	}}
}

func ServiceNameExtract(ctx context.Context) string {
	if v, ok := transport.FromClientContext(ctx); ok {
		res := v.Endpoint()
		if strings.HasPrefix(res, "discovery:///") {
			return strings.TrimPrefix(res, "discovery:///")
		}
	}
	panic("resource name is empty")
}
