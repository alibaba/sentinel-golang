package gin

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
)

// SentinelMiddleware
func SentinelMiddleware(opts ...Option) gin.HandlerFunc {
	options := evaluateOptions(opts)
	return func(c *gin.Context) {
		resourceName := c.Request.Method + ":" + c.Request.URL.Path

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
				c.AbortWithStatus(http.StatusTooManyRequests)
			}
			return
		}

		defer entry.Exit()
		c.Next()
	}
}