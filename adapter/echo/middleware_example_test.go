package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Example() {
	r := echo.New()
	r.Use(middleware.Logger())
	r.Use(
		SentinelMiddleware(
			// customize resource extractor if required
			// method_path by default
			WithResourceExtractor(func(ctx echo.Context) string {
				if res, ok := ctx.Get("X-Real-IP").(string); ok {
					return res
				}
				return ""
			}),
			// customize block fallback if required
			// abort with status 429 by default
			WithBlockFallback(func(ctx echo.Context) error {
				return ctx.JSON(400, map[string]interface{}{
					"err":  "too many requests; the quota used up",
					"code": 10222,
				})
			}),
		),
	)

	r.GET("/test", func(c echo.Context) error {
		return nil
	})
	r.Logger.Fatal(r.Start(":1323"))
}
