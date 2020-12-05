// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	testErrMsg = "test error with caller stack"
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
	return errors.New(testErrMsg)
}

func Test_caller_path(t *testing.T) {
	Error(throwError(), "test error", "k1", "v1")
}

func Test_AssembleMsg(t *testing.T) {
	t.Run("AssembleMsg1", func(t *testing.T) {
		got := AssembleMsg(2, "ERROR", "test msg1", nil, "k1", "v1")
		assert.True(t, strings.Contains(got, `"logLevel":"ERROR","msg":"test msg1","k1":"v1"}`))
	})

	t.Run("AssembleMsg2", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg2", nil, "k1", "v1", "k2", "v2")
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg2","k1":"v1","k2":"v2"}`))
	})

	t.Run("AssembleMsg3", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg3", nil)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg3"}`))
	})

	t.Run("AssembleMsg4", func(t *testing.T) {
		got := AssembleMsg(2, "ERROR", "test msg4", throwError(), "k1", "v1")
		assert.True(t, strings.Contains(got, `"logLevel":"ERROR","msg":"test msg4","k1":"v1"}`))
		assert.True(t, strings.Contains(got, testErrMsg))
	})

	t.Run("AssembleMsg5", func(t *testing.T) {
		got := AssembleMsg(2, "WARN", "test msg5", nil, "reason", throwError())
		assert.True(t, strings.Contains(got, fmt.Sprintf(`"logLevel":"WARN","msg":"test msg5","reason":"%s"`, testErrMsg)))
	})

	t.Run("AssembleMsg6", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg6", nil, "num", 123)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg6","num":123`))
	})

	t.Run("AssembleMsg7", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg7", nil, "num", 123.456)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg7","num":123.456`))
	})

	t.Run("AssembleMsg8", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg8", nil, "flag", true)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg8","flag":true`))
	})

	t.Run("AssembleMsg8", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg8", nil, "flag", true)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg8","flag":true`))
	})

	t.Run("AssembleMsg9", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg9", nil, "object", nil)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg9","object":null`))
	})

	t.Run("AssembleMsg10", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg10", nil, `k1\n\t`, `v1\n\t`)
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg10","k1\n\t":"v1\n\t"`))
	})

	t.Run("AssembleMsg11", func(t *testing.T) {
		got := AssembleMsg(2, "INFO", "test msg11", nil, `json`, "{\"abc\":\"xyz\"}")
		assert.True(t, strings.Contains(got, `"logLevel":"INFO","msg":"test msg11","json":{"abc":"xyz"}`))
	})

}

func Test_caller(t *testing.T) {
	t.Run("caller1", func(t *testing.T) {
		file, _ := caller(1)
		assert.True(t, strings.Contains(file, "logging_test.go"))
	})
}

func TestLogLevelEnabled(t *testing.T) {
	ResetGlobalLoggerLevel(DebugLevel)
	assert.True(t, DebugEnabled(), "Debug should be enabled when log level is DebugLevel")
	assert.True(t, InfoEnabled(), "Info should be enabled when log level is DebugLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is DebugLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is DebugLevel")

	ResetGlobalLoggerLevel(InfoLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is InfoLevel")
	assert.True(t, InfoEnabled(), "Info should be enabled when log level is InfoLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is InfoLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is InfoLevel")

	ResetGlobalLoggerLevel(WarnLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is WarnLevel")
	assert.False(t, InfoEnabled(), "Info should be disabled when log level is WarnLevel")
	assert.True(t, WarnEnabled(), "Warn should be enabled when log level is WarnLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is WarnLevel")

	ResetGlobalLoggerLevel(ErrorLevel)
	assert.False(t, DebugEnabled(), "Debug should be disabled when log level is ErrorLevel")
	assert.False(t, InfoEnabled(), "Info should be disabled when log level is ErrorLevel")
	assert.False(t, WarnEnabled(), "Warn should be disabled when log level is ErrorLevel")
	assert.True(t, ErrorEnabled(), "Error should be enabled when log level is ErrorLevel")
}

func Benchmark_LoggingDebug_Without_Precheck(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	ResetGlobalLoggerLevel(InfoLevel)
	for i := 0; i < b.N; i++ {
		Debug("log test", "k1", "v1", "k2", "v2")
	}
}

func Benchmark_LoggingDebug_With_Precheck(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	ResetGlobalLoggerLevel(InfoLevel)
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
