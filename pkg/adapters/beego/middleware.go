package beego

import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

// SentinelMiddleware returns new web.MiddleWare.
// Default resource name pattern is {httpMethod}:{apiPath}, such as "GET:/api/:id".
// Default block fallback is to return 429 (Too Many Requests) response.
//
// You may customize your own resource extractor and block handler by setting options.
func SentinelMiddleware(opts ...Option) web.MiddleWare {
	options := evaluateOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resourceName := r.Method + ":" + r.URL.Path
			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(r)
			}
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeWeb),
				sentinel.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				if options.blockFallback != nil {
					status, msg := options.blockFallback(r)
					http.Error(w, msg, status)
				} else {
					// default error response
					http.Error(w, "Blocked by Sentinel", http.StatusTooManyRequests)
				}
				return
			}
			defer entry.Exit()
			next.ServeHTTP(w, r)
		})
	}
}
