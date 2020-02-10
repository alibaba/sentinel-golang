package util

import (
	"github.com/sentinel-group/sentinel-golang/logging"
	"testing"
)

func TestWithRecoverGo(t *testing.T) {
	go WithRecoverGo(func() {
		panic("internal error!\n")
	}, logging.GetDefaultLogger())
}