package iris

import (
	"github.com/kataras/iris/v12"
)

type (
	Option  func(*options)
	options struct {
		resourceExtract func(iris.Context) string
		blockFallback   func(iris.Context)
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
func WithResourceExtractor(fn func(iris.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx iris.Context)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
