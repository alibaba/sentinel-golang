package echo

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/labstack/echo/v4"
)

// SentinelMiddleware
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
					err = c.JSON(http.StatusTooManyRequests, "error")
				}
				return err
			}
			defer entry.Exit()
			err = next(c)
			return err
		}

	}
}
