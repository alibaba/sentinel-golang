package goframe

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Example() {
	s := g.Server()
	s.Use(
		SentinelMiddleware(
			// 自定义资源提取器
			WithResourceExtractor(func(r *ghttp.Request) string {
				if res, ok := r.Header["X-Real-IP"]; ok && len(res) > 0 {
					return res[0]
				}
				return ""
			}),
			// 自定义阻塞回退
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
