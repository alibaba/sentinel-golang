package beego

import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"net/http"
)

// SentinelFilterChain returns new web.FilterChain.
// Default resource name pattern is {httpMethod}:{apiPath}, such as "GET:/api/:id".
// Default block fallback is to return 429 (Too Many Requests) response.
//
// You may customize your own resource extractor and block handler by setting options.
func SentinelFilterChain(opts ...Option) web.FilterChain {
	options := evaluateOptions(opts)
	return func(next web.FilterFunc) web.FilterFunc {
		return func(ctx *beegoCtx.Context) {
			resourceName := ctx.Input.Method() + ":" + ctx.Input.URL()
			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(ctx.Request)
			}
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeWeb),
				sentinel.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				if options.blockFallback != nil {
					status, msg := options.blockFallback(ctx.Request)
					http.Error(ctx.ResponseWriter, msg, status)
				} else {
					// default error response
					http.Error(ctx.ResponseWriter, "Blocked by Sentinel", http.StatusTooManyRequests)
				}
				return
			}
			defer entry.Exit()
			next(ctx)
		}
	}
}
