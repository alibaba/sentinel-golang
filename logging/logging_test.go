package logging

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestNewSentinelFileLogger(t *testing.T) {
	fileName := os.TempDir() + "/logger-test.log"
	logger := NewSentinelFileLogger(fileName, "test-log", log.LstdFlags)
	logger.Debug("debug info test.")
	logger.Debugf("debug name is %s", "sim")
	time.Sleep(time.Second * 2)
	_ = os.Remove(fileName)
}
