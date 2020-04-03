package echo

import (
	"github.com/labstack/echo/v4"
)

type (
	Option  func(*options)
	options struct {
		resourceExtract func(echo.Context) string
		blockFallback   func(echo.Context) error
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
func WithResourceExtractor(fn func(ctx echo.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx echo.Context) error) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
