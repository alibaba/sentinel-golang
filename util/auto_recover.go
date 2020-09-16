package util

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

func RunWithRecover(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("Unexpected panic", "err", errors.Errorf("%+v", err))
		}
	}()
	f()
}
