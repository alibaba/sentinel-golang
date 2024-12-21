package goframe

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Example() {
	s := g.Server()
	s.Use(
		SentinelMiddleware(
			// customize resource extractor if required
			WithResourceExtractor(func(r *ghttp.Request) string {
				if res, ok := r.Header["X-Real-IP"]; ok && len(res) > 0 {
					return res[0]
				}
				return ""
			}),
			// customize block fallback if required
			WithBlockFallback(func(r *ghttp.Request) {
				r.Response.WriteHeader(400)
				r.Response.WriteJson(map[string]interface{}{
					"err":  "too many requests; the quota used up",
					"code": 10222,
				})
			}),
		),
	)

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.GET("/test", func(r *ghttp.Request) {
			r.Response.Write("hello sentinel")
		})
	})

	s.SetPort(8199)
}
