package echo

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/labstack/echo/v4"
)

// SentinelMiddleware returns new echo.HandlerFunc.
// Default resource name pattern is {httpMethod}:{apiPath}, such as "GET:/api/:id".
// Default block fallback is to return 429 (Too Many Requests) response.
//
// You may customize your own resource extractor and block handler by setting options.
func SentinelMiddleware(opts ...Option) echo.MiddlewareFunc {
	options := evaluateOptions(opts)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			resourceName := c.Request().Method + ":" + c.Path()
			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(c)
			}
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeWeb),
				sentinel.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				if options.blockFallback != nil {
					err = options.blockFallback(c)
				} else {
					// default error response
					err = c.JSON(http.StatusTooManyRequests, "Blocked by Sentinel")
				}
				return err
			}
			defer entry.Exit()

			err = next(c)
			return err
		}

	}
}
