package hertz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type (
	ServerOption struct {
		Fn func(*serverOptions)
	}
	serverOptions struct {
		resourceExtract func(context.Context, *app.RequestContext) string
		blockFallback   func(context.Context, *app.RequestContext)
	}
)

type (
	ClientOption struct {
		Fn func(o *clientOptions)
	}
	clientOptions struct {
		resourceExtract func(ctx context.Context, req *protocol.Request, resp *protocol.Response) string
		blockFallback   func(ctx context.Context, req *protocol.Request, resp *protocol.Response, blockError error) error
	}
)

func (o *clientOptions) Apply(opts []ClientOption) {
	for _, opt := range opts {
		opt.Fn(o)
	}
}

func (o *serverOptions) Apply(opts []ServerOption) {
	for _, opt := range opts {
		opt.Fn(o)
	}
}

func newServerOptions(opts []ServerOption) *serverOptions {
	options := &serverOptions{
		resourceExtract: func(c context.Context, ctx *app.RequestContext) string {
			return fmt.Sprintf("%v:%v", string(ctx.Request.Method()), ctx.FullPath())
		},
		blockFallback: func(c context.Context, ctx *app.RequestContext) {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		},
	}
	options.Apply(opts)
	return options
}

func newClientOptions(opts []ClientOption) *clientOptions {
	options := &clientOptions{
		resourceExtract: func(ctx context.Context, req *protocol.Request, resp *protocol.Response) string {
			return fmt.Sprintf("%v:%v", string(req.Method()), string(req.Path()))
		},
		blockFallback: func(ctx context.Context, req *protocol.Request, resp *protocol.Response, blockError error) error {
			resp.SetStatusCode(http.StatusTooManyRequests)
			return blockError
		},
	}
	options.Apply(opts)
	return options
}

// WithServerResourceExtractor sets the resource extractor of the web requests for server side.
func WithServerResourceExtractor(fn func(context.Context, *app.RequestContext) string) ServerOption {
	return ServerOption{
		Fn: func(o *serverOptions) {
			o.resourceExtract = fn
		},
	}
}

// WithServerBlockFallback sets the fallback handler when requests are blocked for server side.
func WithServerBlockFallback(fn func(context.Context, *app.RequestContext)) ServerOption {
	return ServerOption{
		Fn: func(o *serverOptions) {
			o.blockFallback = fn
		},
	}
}

// WithClientResourceExtractor sets the resource extractor of the web requests for client side.
func WithClientResourceExtractor(fn func(context.Context, *protocol.Request,
	*protocol.Response) string,
) ClientOption {
	return ClientOption{
		Fn: func(o *clientOptions) {
			o.resourceExtract = fn
		},
	}
}

// WithClientBlockFallback sets the fallback handler when requests are blocked for client side.
func WithClientBlockFallback(fn func(ctx context.Context, req *protocol.Request,
	resp *protocol.Response, blockError error) error,
) ClientOption {
	return ClientOption{
		Fn: func(o *clientOptions) {
			o.blockFallback = fn
		},
	}
}
