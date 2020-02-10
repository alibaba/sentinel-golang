package util

import (
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/logging"
)

func WithRecoverGo(f func(), logger *logging.SentinelLogger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Panic:%+v.", errors.Errorf("%+v", err))
		}
	}()
	f()
}