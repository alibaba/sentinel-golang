package fiber

import "github.com/gofiber/fiber/v2"

func Example() {
	app := fiber.New()
	app.Use(
		SentinelMiddleware(
			// customize resource extractor if required
			// method_path by default
			WithResourceExtractor(func(ctx *fiber.Ctx) string {
				return ctx.GetReqHeaders()["X-Real-IP"]
			}),
			// customize block fallback if required
			// abort with status 429 by default
			WithBlockFallback(func(ctx *fiber.Ctx) error {
				return ctx.Status(400).JSON(struct {
					Error string `json:"error"`
					Code  int    `json:"code"`
				}{
					"too many request; the quota used up",
					10222,
				})
			})),
	)

	app.Get("/test", func(ctx *fiber.Ctx) error { return nil })
	_ = app.Listen(":8080")
}
