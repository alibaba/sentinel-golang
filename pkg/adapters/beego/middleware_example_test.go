package beego

import (
	"github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"net/http"
)

func Example() {
	opts := []Option{
		// customize resource extractor if required
		// method_path by default
		WithResourceExtractor(func(r *http.Request) string {
			return r.Header.Get("X-Real-IP")
		}),
		// customize block fallback if required
		// abort with status 429 by default
		WithBlockFallback(func(r *http.Request) (int, string) {
			return 400, "too many request; the quota used up"
		}),
	}

	web.Get("/test", func(ctx *beegoCtx.Context) {
	})

	// Routing filter chain
	web.InsertFilterChain("/*", SentinelFilterChain(opts...))

	// Global middleware
	web.RunWithMiddleWares(":0", SentinelMiddleware(opts...))
}
