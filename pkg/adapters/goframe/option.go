package goframe

import "github.com/gogf/gf/v2/net/ghttp"

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*ghttp.Request) string
		blockFallback   func(*ghttp.Request)
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
func WithResourceExtractor(fn func(*ghttp.Request) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(r *ghttp.Request)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
