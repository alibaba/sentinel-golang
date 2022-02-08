package go_zero

import (
	"fmt"
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/zeromicro/go-zero/rest"
)

// SentinelMiddleware returns new echo.HandlerFunc.
// Default resource name pattern is {httpMethod}:{apiPath}, such as "GET:/api/:id".
// Default block fallback is to return 429 (Too Many Requests) response.
//
// You may customize your own resource extractor and block handler by setting options.
func SentinelMiddleware(opts ...Option) rest.Middleware {
	fmt.Printf("created\n")
	options := evaluateOptions(opts)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			resourceName := r.Method + ":" + r.URL.Path
			fmt.Printf("%s\n", resourceName)
			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(r)
			}
			fmt.Printf("%s\n", resourceName)
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

			next(w, r)
		}
	}
}
