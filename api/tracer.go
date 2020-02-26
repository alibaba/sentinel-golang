package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

var (
	logger = logging.GetDefaultLogger()
)

func TraceErrorToEntry(entry *base.SentinelEntry, err error, count uint64) {
	if entry == nil {
		return
	}
	TraceErrorToCtx(entry.Context(), err, count)
}

func TraceErrorToCtx(ctx *base.EntryContext, err error, count uint64) {
	defer func() {
		if e := recover(); e != nil {
			logger.Panicf("Fail to execute TraceErrorToCtx, parameter[ctx:%+v, err:%+v, count:%d], sentinel internal error: %+v", ctx, err, count, e)
			return
		}
	}()

	if ctx == nil {
		return
	}
	node := ctx.StatNode
	if node == nil {
		logger.Warnf("The StatNode in EntryContext is nil，ctx:%+v, err:%+v,count:%+v", ctx, err, count)
		return
	}
	traceError(node, err, count)
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
