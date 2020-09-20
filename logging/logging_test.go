package logging

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewSimpleFileLogger(t *testing.T) {
	fileName := "logger-test.log"
	tmpDir := os.TempDir()
	if !strings.HasSuffix(tmpDir, string(os.PathSeparator)) {
		tmpDir = tmpDir + string(os.PathSeparator)
	}
	logger, err := NewSimpleFileLogger(tmpDir + fileName)
	assert.NoError(t, err)

	logger.Info("info test1.")
	logger.Info("info test2.", "name", "sentinel")

	time.Sleep(time.Second * 1)
	_ = os.Remove(fileName)
}

func throwError() error {
	return errors.New("test error with caller stack")
}

func Test_caller_path(t *testing.T) {
	Error(throwError(), "test error", "k1", "v1")
}

func Test_AssembleMsg(t *testing.T) {
	t.Run("AssembleMsg1", func(t *testing.T) {
		got := AssembleMsg(2, "ERROR", "test msg", nil, "k1", "v1")
		assert.True(t, strings.Contains(got, `"logLevel":"ERROR","msg":"test msg","k1":"v1"}`))
	})

	t.Run("AssembleMsg2", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg2", nil, "k1", "v1", "k2", "v2")
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg2","k1":"v1","k2":"v2"}`))
	})

	t.Run("AssembleMsg2", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg2", nil)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg2"}`))
	})

	t.Run("AssembleMsg1", func(t *testing.T) {
		got := AssembleMsg(2, "ERROR", "test msg", throwError(), "k1", "v1")
		assert.True(t, strings.Contains(got, `"logLevel":"ERROR","msg":"test msg","k1":"v1"}`))
		assert.True(t, strings.Contains(got, `test error with caller stack`))
	})
}

func Test_caller(t *testing.T) {
	t.Run("caller1", func(t *testing.T) {
		file, _ := caller(1)
		assert.True(t, strings.Contains(file, "logging_test.go"))
	})
}
