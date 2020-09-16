package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

// TraceError records the provided error to the given SentinelEntry.
func TraceError(entry *base.SentinelEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			logging.Error("failed to TraceError", "err", e)
			return
		}
	}()
	if entry == nil || err == nil {
		return
	}

	entry.SetError(err)
}
