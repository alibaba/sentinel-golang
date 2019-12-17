package util

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestNewSentinelFileLogger(t *testing.T) {
	fileName := "test-logger.log"
	logger := NewSentinelFileLogger(fileName, "test-log", log.LstdFlags|log.Lshortfile)
	logger.Debug("debug info test\n")
	logger.Debugln("debug info test")
	logger.Debugf("debug name is %s", "sim")
	time.Sleep(time.Second * 5)
	_ = os.Remove(fileName)
}
