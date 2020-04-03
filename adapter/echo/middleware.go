package echo

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/labstack/echo/v4"
)

// SentinelMiddleware returns new echo.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/:id"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func SentinelMiddleware(opts ...Option) echo.MiddlewareFunc {

	options := evaluateOptions(opts)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			resourceName := c.Request().Method + ":" + c.Path()
			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(c)
			}
			entry, errEntry := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeWeb),
				sentinel.WithTrafficType(base.Inbound),
			)

			if errEntry != nil {
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
