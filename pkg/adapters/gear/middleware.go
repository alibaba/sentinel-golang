package gear

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/teambition/gear"
)

// SentinelMiddleware returns new gear.Middleware
// Default resource name is {method}:{path}, such as "GET:/api/users/:id"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func SentinelMiddleware(opts ...Option) gear.Middleware {
	options := evaluateOptions(opts)
	return func(ctx *gear.Context) (err error) {
		resourceName := ctx.Method + ":" + gear.GetRouterPatternFromCtx(ctx)

		if options.resourceExtract != nil {
			resourceName = options.resourceExtract(ctx)
		}

		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if blockErr != nil {
			if options.blockFallback != nil {
				err = options.blockFallback(ctx)
			} else {
				err = ctx.End(http.StatusTooManyRequests, []byte("Blocked by Sentinel"))
			}
			return err
		}

		defer entry.Exit()
		return err
	}
}
