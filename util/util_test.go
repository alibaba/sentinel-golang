package util

import (
	"github.com/alibaba/sentinel-golang/logging"
	"testing"
)

func TestWithRecoverGo(t *testing.T) {
	go RunWithRecover(func() {
		panic("internal error!\n")
	}, logging.GetDefaultLogger())
}
