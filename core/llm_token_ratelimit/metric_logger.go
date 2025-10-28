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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

var (
	metricLogger *MetricLogger
	once         sync.Once
)

type MetricLogger struct {
	mu               sync.RWMutex
	writer           *bufio.Writer
	file             *os.File
	baseDir          string
	fileName         string
	currentFileSize  uint64
	maxFileSize      uint64
	maxFileAmount    uint32
	flushIntervalSec uint32
	buffer           []MetricItem
	bufferSize       int
	ticker           *time.Ticker
	stopChan         chan struct{}
	usePid           bool
}

type MetricLoggerConfig struct {
	AppName       string
	LogDir        string
	MaxFileSize   uint64
	MaxFileAmount uint32
	FlushInterval uint32
	UsePid        bool
}

type MetricItem struct {
	Timestamp uint64 `json:"timestamp"`
	RequestID string `json:"request_id"`
	LimitKey  string `json:"limit_key"`

	// PETA.Withhold
	CurrentCapacity    int64 `json:"current_capacity"`
	EstimatedToken     int64 `json:"estimated_token"`
	Difference         int64 `json:"difference"`
	TokenizationLength int   `json:"tokenization_length"`
	WaitingTime        int64 `json:"waiting_time"`

	// PETA.Correct
	ActualToken   int   `json:"actual_token"`
	CorrectResult int64 `json:"correct_result"`
}

func InitMetricLogger(config *MetricLoggerConfig) error {
	var err error
	once.Do(func() {
		metricLogger, err = newMetricLogger(config)
		if err != nil {
			logging.Error(err, "[LLMTokenRateLimit] failed to initialize the MetricLogger")
			return
		}
		metricLogger.startPeriodicFlush()
	})
	return err
}

func FormMetricFileName(config *MetricLoggerConfig) string {
	dot := "."
	separator := "-"
	serviceName := config.AppName

	if strings.Contains(serviceName, dot) {
		serviceName = strings.ReplaceAll(serviceName, dot, separator)
	}
	filename := serviceName + separator + MetricFileNameSuffix + dot + util.FormatDate(util.CurrentTimeMillis())
	if config.UsePid {
		pid := os.Getpid()
		filename = filename + ".pid" + strconv.Itoa(pid)
	}
	return filename
}

func newMetricLogger(config *MetricLoggerConfig) (*MetricLogger, error) {
	logging.Info("[LLMTokenRateLimit] init MetricLogger",
		"config", config.String(),
	)

	logDir := config.LogDir
	if len(logDir) == 0 {
		return nil, fmt.Errorf("log directory cannot be empty")
	}

	maxFileSize := config.MaxFileSize
	if maxFileSize == 0 {
		maxFileSize = DefaultMaxFileSize
	}

	maxFileAmount := config.MaxFileAmount
	if maxFileAmount == 0 {
		maxFileAmount = DefaultMaxFileAmount
	}

	flushInterval := config.FlushInterval
	if flushInterval == 0 {
		flushInterval = DefaultFlushInterval
	}

	if err := util.CreateDirIfNotExists(logDir); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	fileName := FormMetricFileName(config)

	ml := &MetricLogger{
		baseDir:          logDir,
		fileName:         fileName,
		maxFileSize:      maxFileSize,
		maxFileAmount:    maxFileAmount,
		flushIntervalSec: flushInterval,
		bufferSize:       DefaultBufferSize,
		buffer:           make([]MetricItem, 0, DefaultBufferSize),
		stopChan:         make(chan struct{}),
		usePid:           config.UsePid,
	}

	if err := ml.createOrOpenLogFile(); err != nil {
		return nil, err
	}

	return ml, nil
}

func (c *MetricLoggerConfig) String() string {
	return fmt.Sprintf("MetricLoggerConfig{AppName: %s, LogDir: %s, MaxFileSize: %d, MaxFileAmount: %d, FlushInterval: %d, UsePid: %t}",
		c.AppName, c.LogDir, c.MaxFileSize, c.MaxFileAmount, c.FlushInterval, c.UsePid)
}

func (ml *MetricLogger) createOrOpenLogFile() error {
	filePath := filepath.Join(ml.baseDir, ml.fileName)

	if stat, err := os.Stat(filePath); err == nil {
		ml.currentFileSize = uint64(stat.Size())
		if ml.currentFileSize >= ml.maxFileSize {
			if err := ml.rotateLogFile(); err != nil {
				return fmt.Errorf("failed to rotate log file: %v", err)
			}
		}
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	logging.Info("[LLMTokenRateLimit] new metric log file created", "filename", filePath)

	ml.file = file
	ml.writer = bufio.NewWriterSize(file, 8192) // 8KB

	return nil
}

func (ml *MetricLogger) rotateLogFile() error {
	if ml.writer != nil {
		ml.writer.Flush()
		ml.writer = nil
	}
	if ml.file != nil {
		ml.file.Close()
		ml.file = nil
	}

	currentPath := filepath.Join(ml.baseDir, ml.fileName)

	for i := int(ml.maxFileAmount) - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", currentPath, i)
		newPath := fmt.Sprintf("%s.%d", currentPath, i+1)

		if i == int(ml.maxFileAmount)-1 {
			if err := os.Remove(newPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("[LLMTokenRateLimit] failed to remove old log file %s: %v",
					newPath, err)
			}
		}

		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to rotate log file from %s to %s: %v",
					oldPath, newPath, err)
			}
		}
	}

	if _, err := os.Stat(currentPath); err == nil {
		if err := os.Rename(currentPath, currentPath+".1"); err != nil {
			return fmt.Errorf("failed to rotate current log file from %s to %s: %v",
				currentPath, currentPath+".1", err)
		}
	}

	ml.currentFileSize = 0

	return ml.createOrOpenLogFile()
}

func (ml *MetricLogger) startPeriodicFlush() {
	ml.ticker = time.NewTicker(time.Duration(ml.flushIntervalSec) * time.Second)
	go func() {
		for {
			select {
			case <-ml.ticker.C:
				ml.flush()
			case <-ml.stopChan:
				ml.ticker.Stop()
				ml.flush()
				return
			}
		}
	}()
}

func (ml *MetricLogger) Record(item MetricItem) {
	if ml == nil {
		return
	}

	if item.Timestamp == 0 {
		item.Timestamp = util.CurrentTimeMillis()
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.buffer = append(ml.buffer, item)

	if len(ml.buffer) > ml.bufferSize {
		ml.flushUnsafe()
	}
}

func (ml *MetricLogger) flush() {
	if ml == nil {
		return
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.flushUnsafe()
}

func (ml *MetricLogger) flushUnsafe() {
	if len(ml.buffer) == 0 || ml.writer == nil {
		return
	}

	for _, item := range ml.buffer {
		line := ml.formatLogLine(item)
		n, err := ml.writer.WriteString(line)
		if err != nil {
			logging.Error(err, "[LLMTokenRateLimit] failed to write log line")
			continue
		}

		ml.currentFileSize += uint64(n)

		if ml.currentFileSize >= ml.maxFileSize {
			ml.writer.Flush()
			if err := ml.rotateLogFile(); err != nil {
				logging.Error(err, "[LLMTokenRateLimit] failed to rotate log file")
			}
		}
	}

	ml.writer.Flush()
	ml.buffer = ml.buffer[:0]
}

func (ml *MetricLogger) formatLogLine(item MetricItem) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%d|%d|%d|%d|%d\n",
		util.FormatTimeMillis(item.Timestamp),
		item.RequestID,
		item.LimitKey,
		item.CurrentCapacity,
		item.WaitingTime,
		item.EstimatedToken,
		item.Difference,
		item.TokenizationLength,
		item.ActualToken,
		item.CorrectResult,
	)
}

func (ml *MetricLogger) Stop() {
	if ml == nil {
		return
	}

	close(ml.stopChan)

	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.flushUnsafe()

	if ml.writer != nil {
		ml.writer = nil
	}

	if ml.file != nil {
		ml.file.Close()
		ml.file = nil
	}
}

func RecordMetric(item MetricItem) {
	if metricLogger != nil {
		metricLogger.Record(item)
	}
}

func StopMetricLogger() {
	if metricLogger != nil {
		metricLogger.Stop()
	}
}
