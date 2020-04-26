package util

import (
	"testing"

	"github.com/alibaba/sentinel-golang/logging"
)

func TestWithRecoverGo(t *testing.T) {
	go RunWithRecover(func() {
		panic("internal error!\n")
	}, logging.GetDefaultLogger())
}
