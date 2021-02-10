package gear

import (
	"github.com/teambition/gear"
)

func Example() {
	app := gear.New()

	router := gear.NewRouter()
	router.Use(
		SentinelMiddleware(
			// customize resource extractor if required
			// method_path by default
			WithResourceExtractor(func(ctx *gear.Context) string {
				return ctx.GetHeader("X-Real-IP")
			}),
			// customize block fallback if required
			// abort with status 429 by default
			WithBlockFallback(func(ctx *gear.Context) error {
				return ctx.JSON(400, map[string]interface{}{
					"err":  "too many request; the quota used up",
					"code": 10222,
				})
			}),
		),
	)
	router.Get("/test", func(c *gear.Context) error {
		return nil
	})
	app.UseHandler(router)
	_ = app.Listen(":0")
}
