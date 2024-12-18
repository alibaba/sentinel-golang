package goframe

import (
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gogf/gf/v2/net/ghttp"
	"net/http"
)

// SentinelMiddleware returns new ghttp.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/:id"
// Default block fallback is returning 429 status code
// Define your own behavior by setting options
func SentinelMiddleware(opts ...Option) ghttp.HandlerFunc {
	options := evaluateOptions(opts)
	return func(r *ghttp.Request) {
		resourceName := r.Method + ":" + r.URL.Path

		if options.resourceExtract != nil {
			extractedName := options.resourceExtract(r)
			if extractedName == "" {
				extractedName = resourceName
			}
			resourceName = extractedName
		}

		entry, err := api.Entry(
			resourceName,
			api.WithResourceType(base.ResTypeWeb),
			api.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				options.blockFallback(r)
			} else {
				r.Response.WriteHeader(http.StatusTooManyRequests)
				r.Response.Writeln("Too Many Requests")
			}
			return
		}

		defer entry.Exit()

		r.Middleware.Next()
	}
}
