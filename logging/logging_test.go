package logging

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleFileLogger(t *testing.T) {
	fileName := "logger-test.log"
	tmpDir := os.TempDir()
	if !strings.HasSuffix(tmpDir, string(os.PathSeparator)) {
		tmpDir = tmpDir + string(os.PathSeparator)
	}
	logger, err := NewSimpleFileLogger(tmpDir+fileName, "test-log", log.LstdFlags|log.LstdFlags)
	assert.NoError(t, err)

	logger.Debug("debug info test.")
	logger.Infof("Hello %s", "sentinel")

	time.Sleep(time.Second * 1)
	_ = os.Remove(fileName)
}
