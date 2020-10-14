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

func TestLogLevelEnabled(t *testing.T) {
	SetGlobalLoggerLevel(DebugLevel)
	assert.True(t, DebugEnabled(), "Debug should be enabled when log level is DebugLevel")
	assert.True(t, InfoEnabled(), "Info should be enabled when log level is DebugLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is DebugLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is DebugLevel")

	SetGlobalLoggerLevel(InfoLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is InfoLevel")
	assert.True(t, InfoEnabled(), "Info should be enabled when log level is InfoLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is InfoLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is InfoLevel")

	SetGlobalLoggerLevel(WarnLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is WarnLevel")
	assert.False(t, InfoEnabled(), "Info should be disabled when log level is WarnLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is WarnLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is WarnLevel")

	SetGlobalLoggerLevel(ErrorLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is ErrorLevel")
	assert.False(t, InfoEnabled(), "Info should be disabled when log level is ErrorLevel")
	assert.False(t, WarnEnabled(), "Warn should be disabled when log level is ErrorLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is ErrorLevel")
}

func Benchmark_LoggingDebug_Without_Precheck(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	SetGlobalLoggerLevel(InfoLevel)
	for i := 0; i < b.N; i++ {
		Debug("log test", "k1", "v1", "k2", "v2")
	}
}

func Benchmark_LoggingDebug_With_Precheck(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	SetGlobalLoggerLevel(InfoLevel)
	for i := 0; i < b.N; i++ {
		if DebugEnabled() {
			Debug("log test", "k1", "v1", "k2", "v2")
		}
	}
}

func BenchmarkAssembleMsg(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AssembleMsg(1, "INFO", "test msg", nil, "k1", "v1", "k2", "v2")
	}
}
