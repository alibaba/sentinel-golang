package util

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

func RunWithRecover(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error(errors.Errorf("%+v", err), "unexpected panic")
		}
	}()
	f()
}
