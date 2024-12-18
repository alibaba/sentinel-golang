package goframe

import "github.com/gogf/gf/v2/net/ghttp"

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*ghttp.Request) string
		blockFallback   func(*ghttp.Request)
	}
)

// evaluateOptions 评估选项
func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}

	return optCopy
}

// WithResourceExtractor 设置资源提取器
func WithResourceExtractor(fn func(*ghttp.Request) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback 设置被流控的回退处理函数
func WithBlockFallback(fn func(r *ghttp.Request)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
