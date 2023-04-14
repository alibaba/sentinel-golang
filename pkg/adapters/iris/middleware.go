package iris

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/kataras/iris/v12"
)

func SentinelMiddleware(opts ...Option) iris.Handler {
	options := evaluateOptions(opts)
	return func(c iris.Context) {
		resourceName := c.Request().Method + ":" + c.Request().URL.String()

		if options.resourceExtract != nil {
			resourceName = options.resourceExtract(c)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				options.blockFallback(c)
			} else {
				c.StatusCode(http.StatusTooManyRequests)
				c.StopExecution()
			}
			return
		}

		defer entry.Exit()
		c.Next()
	}
}
