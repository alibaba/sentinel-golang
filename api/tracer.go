package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

var (
	logger = logging.GetDefaultLogger()
)

type TraceErrorOptions struct {
	count uint64
}

type TraceErrorOption func(*TraceErrorOptions)

func WithCount(count uint64) TraceErrorOption {
	return func(opts *TraceErrorOptions) {
		opts.count = count
	}
}

func TraceErrorToEntry(entry *base.SentinelEntry, err error, opts ...TraceErrorOption) {
	if entry == nil {
		return
	}

	TraceErrorToCtx(entry.Context(), err, opts...)
}

func TraceErrorToCtx(ctx *base.EntryContext, err error, opts ...TraceErrorOption) {
	defer func() {
		if e := recover(); e != nil {
			logger.Panicf("Fail to execute TraceErrorToCtx, parameter[ctx:%+v, err:%+v, opts:%+v], sentinel internal error: %+v", ctx, err, opts, e)
			return
		}
	}()

	if ctx == nil {
		return
	}
	node := ctx.StatNode
	if node == nil {
		logger.Warnf("The StatNode in EntryContext is nil，ctx:%+v, err:%+v,opt:%+v", ctx, err, opts)
		return
	}

	var options = TraceErrorOptions{
		count: 1,
	}
	for _, opt := range opts {
		opt(&options)
	}

	traceError(node, err, options.count)
}

func traceError(node base.StatNode, err error, cnt uint64) {
	if node == nil {
		return
	}
	if cnt <= 0 {
		return
	}
	node.AddMetric(base.MetricEventError, uint64(cnt))
}
