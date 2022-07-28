package hertz

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/cloudwego/hertz/pkg/app"
)

// SentinelServerMiddleware returns new app.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/:id"
// Default block fallback is returning 429 code
// Define your own behavior by setting serverOptions
func SentinelServerMiddleware(opts ...ServerOption) app.HandlerFunc {
	options := newServerOptions(opts)
	return func(c context.Context, ctx *app.RequestContext) {
		resourceName := options.resourceExtract(c, ctx)

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)
		if err != nil {
			options.blockFallback(c, ctx)
			return
		}
		defer entry.Exit()
		ctx.Next(c)
	}
}
