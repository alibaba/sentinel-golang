package gear

import (
	"github.com/teambition/gear"
)

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*gear.Context) string
		blockFallback   func(*gear.Context) error
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
func WithResourceExtractor(fn func(*gear.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx *gear.Context) error) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
