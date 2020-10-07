package api

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

// TraceError records the provided error to the given SentinelEntry.
func TraceError(entry *base.SentinelEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			logging.Error(errors.Errorf("%+v", e), "failed to TraceError")
			return
		}
	}()
	if entry == nil || err == nil {
		return
	}

	entry.SetError(err)
}
