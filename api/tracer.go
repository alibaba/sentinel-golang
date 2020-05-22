package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

// TraceError records the provided error to the given SentinelEntry.
func TraceError(entry *base.SentinelEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			logging.GetDefaultLogger().Panicf("Failed to TraceError, panic error: %+v", e)
			return
		}
	}()
	if entry == nil || err == nil {
		return
	}

	entry.SetError(err)
}

func traceErrorToNode(node base.StatNode, err error, cnt uint64) {
	if node == nil {
		return
	}
	if cnt <= 0 {
		return
	}
	node.AddMetric(base.MetricEventError, cnt)
}
