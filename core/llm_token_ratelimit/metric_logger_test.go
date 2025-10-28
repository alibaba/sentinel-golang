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

package llm_token_ratelimit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/alibaba/sentinel-golang/util"
)

// Test constants and helpers
const (
	TestLogDir        = "./test_logs"
	TestAppName       = "test-app"
	TestMaxFileSize   = 1024 // 1KB for testing
	TestMaxFileAmount = 3
	TestFlushInterval = 1
)

// Helper function to create test config
func createTestConfig() *MetricLoggerConfig {
	return &MetricLoggerConfig{
		AppName:       TestAppName,
		LogDir:        TestLogDir,
		MaxFileSize:   TestMaxFileSize,
		MaxFileAmount: TestMaxFileAmount,
		FlushInterval: TestFlushInterval,
		UsePid:        false,
	}
}

// Helper function to clean up test directory
func cleanupTestDir(t *testing.T) {
	if err := os.RemoveAll(TestLogDir); err != nil {
		t.Logf("Failed to remove test directory: %v", err)
	}
}

// Helper function to create test metric item
func createTestMetricItem() MetricItem {
	return MetricItem{
		Timestamp:          util.CurrentTimeMillis(),
		RequestID:          "test-request-123",
		LimitKey:           "test-limit-key",
		CurrentCapacity:    1000,
		EstimatedToken:     50,
		Difference:         -50,
		TokenizationLength: 25,
		WaitingTime:        100,
		ActualToken:        45,
		CorrectResult:      0,
	}
}

// Test MetricLoggerConfig.String function
func TestMetricLoggerConfig_String(t *testing.T) {
	tests := []struct {
		name   string
		config *MetricLoggerConfig
		expect string
	}{
		{
			name: "Complete config",
			config: &MetricLoggerConfig{
				AppName:       "test-app",
				LogDir:        "/logs",
				MaxFileSize:   1024,
				MaxFileAmount: 5,
				FlushInterval: 2,
				UsePid:        true,
			},
			expect: "MetricLoggerConfig{AppName: test-app, LogDir: /logs, MaxFileSize: 1024, MaxFileAmount: 5, FlushInterval: 2, UsePid: true}",
		},
		{
			name: "Empty config",
			config: &MetricLoggerConfig{
				AppName:       "",
				LogDir:        "",
				MaxFileSize:   0,
				MaxFileAmount: 0,
				FlushInterval: 0,
				UsePid:        false,
			},
			expect: "MetricLoggerConfig{AppName: , LogDir: , MaxFileSize: 0, MaxFileAmount: 0, FlushInterval: 0, UsePid: false}",
		},
		{
			name: "Config with special characters",
			config: &MetricLoggerConfig{
				AppName:       "app-with.dots",
				LogDir:        "/path/with spaces",
				MaxFileSize:   999999,
				MaxFileAmount: 10,
				FlushInterval: 5,
				UsePid:        true,
			},
			expect: "MetricLoggerConfig{AppName: app-with.dots, LogDir: /path/with spaces, MaxFileSize: 999999, MaxFileAmount: 10, FlushInterval: 5, UsePid: true}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

// Test FormMetricFileName function
func TestFormMetricFileName(t *testing.T) {
	tests := []struct {
		name   string
		config *MetricLoggerConfig
		verify func(string) bool
	}{
		{
			name: "Basic app name without dots",
			config: &MetricLoggerConfig{
				AppName: "testapp",
				UsePid:  false,
			},
			verify: func(filename string) bool {
				return strings.Contains(filename, "testapp-"+MetricFileNameSuffix) &&
					strings.Contains(filename, util.FormatDate(util.CurrentTimeMillis())) &&
					!strings.Contains(filename, ".pid")
			},
		},
		{
			name: "App name with dots (should be replaced with dashes)",
			config: &MetricLoggerConfig{
				AppName: "test.app.name",
				UsePid:  false,
			},
			verify: func(filename string) bool {
				return strings.Contains(filename, "test-app-name-"+MetricFileNameSuffix) &&
					!strings.Contains(filename, "test.app.name")
			},
		},
		{
			name: "With PID enabled",
			config: &MetricLoggerConfig{
				AppName: "testapp",
				UsePid:  true,
			},
			verify: func(filename string) bool {
				pid := os.Getpid()
				expectedPidSuffix := ".pid" + strconv.Itoa(pid)
				return strings.Contains(filename, "testapp-"+MetricFileNameSuffix) &&
					strings.Contains(filename, expectedPidSuffix)
			},
		},
		{
			name: "Empty app name",
			config: &MetricLoggerConfig{
				AppName: "",
				UsePid:  false,
			},
			verify: func(filename string) bool {
				return strings.HasPrefix(filename, "-"+MetricFileNameSuffix)
			},
		},
		{
			name: "App name with multiple consecutive dots",
			config: &MetricLoggerConfig{
				AppName: "test..app...name",
				UsePid:  false,
			},
			verify: func(filename string) bool {
				return strings.Contains(filename, "test--app---name-"+MetricFileNameSuffix)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := FormMetricFileName(tt.config)

			if filename == "" {
				t.Error("Expected non-empty filename, got empty string")
			}

			if !tt.verify(filename) {
				t.Errorf("Filename %q did not meet verification criteria", filename)
			}
		})
	}
}

// Test InitMetricLogger function
func TestInitMetricLogger(t *testing.T) {
	defer cleanupTestDir(t)

	// Reset global state
	metricLogger = nil
	once = sync.Once{}

	tests := []struct {
		name      string
		config    *MetricLoggerConfig
		expectErr bool
	}{
		{
			name:      "Valid config",
			config:    createTestConfig(),
			expectErr: false,
		},
		{
			name: "Empty log directory",
			config: &MetricLoggerConfig{
				AppName:       TestAppName,
				LogDir:        "",
				MaxFileSize:   TestMaxFileSize,
				MaxFileAmount: TestMaxFileAmount,
				FlushInterval: TestFlushInterval,
			},
			expectErr: true,
		},
		{
			name: "Config with default values",
			config: &MetricLoggerConfig{
				AppName: TestAppName,
				LogDir:  TestLogDir,
				// Other fields will use defaults
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state for each test
			metricLogger = nil
			once = sync.Once{}

			err := InitMetricLogger(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if metricLogger == nil {
					t.Error("Expected metricLogger to be initialized, got nil")
				}
			}

			// Clean up after each test
			if metricLogger != nil {
				metricLogger.Stop()
			}
			cleanupTestDir(t)
		})
	}
}

// Test InitMetricLogger multiple calls (should only initialize once)
func TestInitMetricLogger_MultipleCallsOnlyInitializeOnce(t *testing.T) {
	defer cleanupTestDir(t)

	// Reset global state
	metricLogger = nil
	once = sync.Once{}

	config := createTestConfig()

	// First call
	err1 := InitMetricLogger(config)
	if err1 != nil {
		t.Fatalf("First initialization failed: %v", err1)
	}

	firstLogger := metricLogger

	// Second call
	err2 := InitMetricLogger(config)
	if err2 != nil {
		t.Errorf("Second call should not return error, got %v", err2)
	}

	// Should be the same instance
	if metricLogger != firstLogger {
		t.Error("Expected same logger instance on multiple calls")
	}

	// Clean up
	if metricLogger != nil {
		metricLogger.Stop()
	}
}

// Test newMetricLogger function
func TestNewMetricLogger(t *testing.T) {
	defer cleanupTestDir(t)

	tests := []struct {
		name      string
		config    *MetricLoggerConfig
		expectErr bool
		verify    func(*MetricLogger) bool
	}{
		{
			name:      "Valid config",
			config:    createTestConfig(),
			expectErr: false,
			verify: func(ml *MetricLogger) bool {
				return ml.baseDir == TestLogDir &&
					ml.maxFileSize == TestMaxFileSize &&
					ml.maxFileAmount == TestMaxFileAmount &&
					ml.flushIntervalSec == TestFlushInterval &&
					ml.bufferSize == DefaultBufferSize
			},
		},
		{
			name: "Config with zero values (should use defaults)",
			config: &MetricLoggerConfig{
				AppName: TestAppName,
				LogDir:  TestLogDir,
				// Other fields are zero, should use defaults
			},
			expectErr: false,
			verify: func(ml *MetricLogger) bool {
				return ml.maxFileSize == DefaultMaxFileSize &&
					ml.maxFileAmount == DefaultMaxFileAmount &&
					ml.flushIntervalSec == DefaultFlushInterval
			},
		},
		{
			name: "Empty log directory",
			config: &MetricLoggerConfig{
				AppName:       TestAppName,
				LogDir:        "",
				MaxFileSize:   TestMaxFileSize,
				MaxFileAmount: TestMaxFileAmount,
				FlushInterval: TestFlushInterval,
			},
			expectErr: true,
			verify:    nil,
		},
		{
			name: "Invalid log directory path",
			config: &MetricLoggerConfig{
				AppName:       TestAppName,
				LogDir:        "/invalid/\x00/path", // Contains null byte
				MaxFileSize:   TestMaxFileSize,
				MaxFileAmount: TestMaxFileAmount,
				FlushInterval: TestFlushInterval,
			},
			expectErr: true,
			verify:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml, err := newMetricLogger(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if ml != nil {
					t.Error("Expected nil MetricLogger on error, got non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if ml == nil {
					t.Error("Expected non-nil MetricLogger, got nil")
				} else {
					if tt.verify != nil && !tt.verify(ml) {
						t.Error("MetricLogger verification failed")
					}
					ml.Stop()
				}
			}

			cleanupTestDir(t)
		})
	}
}

// Test MetricLogger.createOrOpenLogFile function
func TestMetricLogger_CreateOrOpenLogFile(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	// Verify file was created
	filePath := filepath.Join(ml.baseDir, ml.fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected log file to be created, but it doesn't exist")
	}

	// Verify file handles are set
	if ml.file == nil {
		t.Error("Expected file handle to be set, got nil")
	}
	if ml.writer == nil {
		t.Error("Expected writer to be set, got nil")
	}
}

// Test MetricLogger.Record function
func TestMetricLogger_Record(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	tests := []struct {
		name string
		item MetricItem
	}{
		{
			name: "Complete metric item",
			item: createTestMetricItem(),
		},
		{
			name: "Metric item with zero timestamp (should be set automatically)",
			item: MetricItem{
				RequestID:          "test-request-456",
				LimitKey:           "test-limit-key-2",
				CurrentCapacity:    500,
				EstimatedToken:     25,
				Difference:         -25,
				TokenizationLength: 12,
				WaitingTime:        50,
				ActualToken:        20,
				CorrectResult:      1,
			},
		},
		{
			name: "Metric item with negative values",
			item: MetricItem{
				Timestamp:          util.CurrentTimeMillis(),
				RequestID:          "test-request-789",
				LimitKey:           "test-limit-key-3",
				CurrentCapacity:    -100,
				EstimatedToken:     -10,
				Difference:         10,
				TokenizationLength: 5,
				WaitingTime:        0,
				ActualToken:        -5,
				CorrectResult:      2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialBufferLen := len(ml.buffer)

			ml.Record(tt.item)

			// Check if item was added to buffer
			if len(ml.buffer) != initialBufferLen+1 {
				t.Errorf("Expected buffer length to increase by 1, got %d -> %d",
					initialBufferLen, len(ml.buffer))
			}

			// Check if timestamp was set for zero timestamp
			if tt.item.Timestamp == 0 {
				lastItem := ml.buffer[len(ml.buffer)-1]
				if lastItem.Timestamp == 0 {
					t.Error("Expected timestamp to be set automatically, got 0")
				}
			}
		})
	}
}

// Test MetricLogger.Record with nil logger
func TestMetricLogger_Record_NilLogger(t *testing.T) {
	var ml *MetricLogger = nil
	item := createTestMetricItem()

	// Should not panic
	ml.Record(item)
}

// Test MetricLogger.Record buffer overflow
func TestMetricLogger_Record_BufferOverflow(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	// Fill buffer to capacity
	for i := 0; i < DefaultBufferSize; i++ {
		item := createTestMetricItem()
		item.RequestID = fmt.Sprintf("request-%d", i)
		ml.Record(item)
	}

	// Buffer should be full but not flushed yet
	if len(ml.buffer) != DefaultBufferSize {
		t.Errorf("Expected buffer length to be %d, got %d", DefaultBufferSize, len(ml.buffer))
	}

	// Add one more item, should trigger flush
	overflowItem := createTestMetricItem()
	overflowItem.RequestID = "overflow-request"
	ml.Record(overflowItem)

	// Buffer should be empty after flush
	if len(ml.buffer) != 0 {
		t.Errorf("Expected buffer to be empty after overflow flush, got length %d", len(ml.buffer))
	}
}

// Test MetricLogger.flush function
func TestMetricLogger_Flush(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	// Add some items to buffer
	for i := 0; i < 5; i++ {
		item := createTestMetricItem()
		item.RequestID = fmt.Sprintf("request-%d", i)
		ml.Record(item)
	}

	// Verify buffer is not empty
	if len(ml.buffer) == 0 {
		t.Error("Expected buffer to have items before flush")
	}

	// Flush
	ml.flush()

	// Verify buffer is empty after flush
	if len(ml.buffer) != 0 {
		t.Errorf("Expected buffer to be empty after flush, got length %d", len(ml.buffer))
	}

	// Verify data was written to file
	filePath := filepath.Join(ml.baseDir, ml.fileName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected log file to contain data after flush, got empty file")
	}

	// Verify log format
	lines := strings.Split(string(content), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
			// Check if line contains expected separators
			if !strings.Contains(line, "|") {
				t.Errorf("Expected log line to contain '|' separators, got: %s", line)
			}
		}
	}

	if nonEmptyLines != 5 {
		t.Errorf("Expected 5 log lines, got %d", nonEmptyLines)
	}
}

// Test MetricLogger.flush with nil logger
func TestMetricLogger_Flush_NilLogger(t *testing.T) {
	var ml *MetricLogger = nil

	// Should not panic
	ml.flush()
}

// Test MetricLogger.formatLogLine function
func TestMetricLogger_FormatLogLine(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	item := MetricItem{
		Timestamp:          1234567890123,
		RequestID:          "test-request",
		LimitKey:           "test-key",
		CurrentCapacity:    1000,
		EstimatedToken:     50,
		Difference:         -50,
		TokenizationLength: 25,
		WaitingTime:        100,
		ActualToken:        45,
		CorrectResult:      0,
	}

	line := ml.formatLogLine(item)

	// Verify line is not empty
	if line == "" {
		t.Error("Expected non-empty log line, got empty string")
	}

	// Verify line ends with newline
	if !strings.HasSuffix(line, "\n") {
		t.Error("Expected log line to end with newline")
	}

	// Verify line contains expected separators
	separatorCount := strings.Count(line, "|")
	expectedSeparators := 9 // Based on the format string
	if separatorCount != expectedSeparators {
		t.Errorf("Expected %d separators, got %d", expectedSeparators, separatorCount)
	}

	// Verify specific fields are present
	expectedFields := []string{
		"test-request",
		"test-key",
		"1000",
		"100",
		"50",
		"-50",
		"25",
		"45",
		"0",
	}

	for _, field := range expectedFields {
		if !strings.Contains(line, field) {
			t.Errorf("Expected log line to contain %q, got: %s", field, line)
		}
	}
}

// Test MetricLogger.rotateLogFile function
func TestMetricLogger_RotateLogFile(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	config.MaxFileAmount = 3 // Keep it small for testing

	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	// Create initial log file with some content
	initialContent := "initial log content\n"
	filePath := filepath.Join(ml.baseDir, ml.fileName)

	if err := ioutil.WriteFile(filePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	// Force rotation
	err = ml.rotateLogFile()
	if err != nil {
		t.Errorf("Expected no error during rotation, got %v", err)
	}

	// Verify original file was rotated to .1
	rotatedPath := filePath + ".1"
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Error("Expected rotated file to exist at .1 suffix")
	}

	// Verify new file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected new log file to be created after rotation")
	}

	// Verify rotated file contains original content
	rotatedContent, err := ioutil.ReadFile(rotatedPath)
	if err != nil {
		t.Fatalf("Failed to read rotated file: %v", err)
	}

	if string(rotatedContent) != initialContent {
		t.Errorf("Expected rotated file to contain %q, got %q", initialContent, string(rotatedContent))
	}

	// Verify current file size was reset
	if ml.currentFileSize != 0 {
		t.Errorf("Expected currentFileSize to be reset to 0, got %d", ml.currentFileSize)
	}
}

// Test MetricLogger.Stop function
func TestMetricLogger_Stop(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}

	// Start periodic flush to have ticker running
	ml.startPeriodicFlush()

	// Add some items to buffer
	for i := 0; i < 3; i++ {
		item := createTestMetricItem()
		item.RequestID = fmt.Sprintf("stop-test-request-%d", i)
		ml.Record(item)
	}

	// Stop the logger
	ml.Stop()

	// Verify resources are cleaned up
	if ml.writer != nil {
		t.Error("Expected writer to be nil after stop")
	}
	if ml.file != nil {
		t.Error("Expected file to be nil after stop")
	}

	// Verify final flush occurred (buffer should be empty)
	if len(ml.buffer) != 0 {
		t.Errorf("Expected buffer to be empty after stop, got length %d", len(ml.buffer))
	}
}

// Test MetricLogger.Stop with nil logger
func TestMetricLogger_Stop_NilLogger(t *testing.T) {
	var ml *MetricLogger = nil

	// Should not panic
	ml.Stop()
}

// Test RecordMetric function
func TestRecordMetric(t *testing.T) {
	defer cleanupTestDir(t)

	// Reset global state
	metricLogger = nil
	once = sync.Once{}

	// Initialize global logger
	config := createTestConfig()
	err := InitMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize global metric logger: %v", err)
	}
	defer StopMetricLogger()

	item := createTestMetricItem()

	// Record metric
	RecordMetric(item)

	// Verify item was recorded
	if len(metricLogger.buffer) == 0 {
		t.Error("Expected buffer to contain recorded item, got empty buffer")
	}
}

// Test RecordMetric with nil global logger
func TestRecordMetric_NilGlobalLogger(t *testing.T) {
	// Ensure global logger is nil
	metricLogger = nil

	item := createTestMetricItem()

	// Should not panic
	RecordMetric(item)
}

// Test StopMetricLogger function
func TestStopMetricLogger(t *testing.T) {
	defer cleanupTestDir(t)

	// Reset global state
	metricLogger = nil
	once = sync.Once{}

	// Initialize global logger
	config := createTestConfig()
	err := InitMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize global metric logger: %v", err)
	}

	// Add some items
	for i := 0; i < 3; i++ {
		item := createTestMetricItem()
		item.RequestID = fmt.Sprintf("global-stop-request-%d", i)
		RecordMetric(item)
	}

	// Stop global logger
	StopMetricLogger()

	// Verify logger was stopped (resources cleaned up)
	if metricLogger.writer != nil {
		t.Error("Expected global logger writer to be nil after stop")
	}
	if metricLogger.file != nil {
		t.Error("Expected global logger file to be nil after stop")
	}
}

// Test StopMetricLogger with nil global logger
func TestStopMetricLogger_NilGlobalLogger(t *testing.T) {
	// Ensure global logger is nil
	metricLogger = nil

	// Should not panic
	StopMetricLogger()
}

// Test concurrent access to MetricLogger
func TestMetricLogger_ConcurrentAccess(t *testing.T) {
	defer cleanupTestDir(t)

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		t.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	const numGoroutines = 6
	const itemsPerGoroutine = 9

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines writing concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				item := createTestMetricItem()
				item.RequestID = fmt.Sprintf("concurrent-%d-%d", goroutineID, j)
				ml.Record(item)
			}
		}(i)
	}

	wg.Wait()

	// Flush to ensure all data is written
	ml.flush()

	// Verify data was written to file
	totalLines := 0

	// Count lines in main log file
	mainFilePath := filepath.Join(ml.baseDir, ml.fileName)
	if content, err := ioutil.ReadFile(mainFilePath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				totalLines++
			}
		}
	}

	// Count lines in rotated log files (.1, .2, .3, etc.)
	for i := 1; i <= int(TestMaxFileAmount); i++ {
		rotatedFilePath := fmt.Sprintf("%s.%d", mainFilePath, i)
		if content, err := ioutil.ReadFile(rotatedFilePath); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					totalLines++
				}
			}
			t.Logf("Found rotated file %s with %d lines", rotatedFilePath,
				len(strings.Split(string(content), "\n"))-1) // -1 for empty last line
		}
	}

	expectedLines := numGoroutines * itemsPerGoroutine
	if totalLines != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, totalLines)
	}
}

// Benchmark tests
func BenchmarkMetricLogger_Record(b *testing.B) {
	defer cleanupTestDir(&testing.T{})

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		b.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	item := createTestMetricItem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item.RequestID = fmt.Sprintf("bench-request-%d", i)
		ml.Record(item)
	}
}

func BenchmarkMetricLogger_FormatLogLine(b *testing.B) {
	defer cleanupTestDir(&testing.T{})

	config := createTestConfig()
	ml, err := newMetricLogger(config)
	if err != nil {
		b.Fatalf("Failed to create MetricLogger: %v", err)
	}
	defer ml.Stop()

	item := createTestMetricItem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ml.formatLogLine(item)
	}
}

func BenchmarkFormMetricFileName(b *testing.B) {
	config := createTestConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormMetricFileName(config)
	}
}
