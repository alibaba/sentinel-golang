package util

import (
	"github.com/pkg/errors"
	"github.com/alibaba/sentinel-golang/logging"
)

func RunWithRecover(f func(), logger logging.Logger) {
	defer func() {
		if err := recover(); err != nil && logger != nil {
			logger.Panicf("Unexpected panic: %+v", errors.Errorf("%+v", err))
		}
	}()
	f()
}
