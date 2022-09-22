package fiber

import "github.com/gofiber/fiber/v2"

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*fiber.Ctx) string
		blockFallback   func(*fiber.Ctx) error
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}

	return optCopy
}

// WithResourceExtractor sets the resource extractor of the web requests.
func WithResourceExtractor(fn func(*fiber.Ctx) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx *fiber.Ctx) error) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
