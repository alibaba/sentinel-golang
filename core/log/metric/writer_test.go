package metric

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/alibaba/sentinel-golang/util"
)

func TestDefaultMetricLogWriter_rotateWithDateAndNewFileName(t *testing.T) {
	_, offset := util.Now().Zone()
	fmt.Println("old offset:", offset)
	type fields struct {
		baseDir           string
		baseFilename      string
		maxSingleSize     uint64
		maxFileAmount     uint32
		timezoneOffsetSec int64
		latestOpSec       int64
		curMetricFile     *os.File
		curMetricIdxFile  *os.File
		metricOut         *bufio.Writer
		idxOut            *bufio.Writer
		mux               *sync.RWMutex
	}
	type args struct {
		time   uint64
		before func()
		file   func()
		clear  func()
	}
	baseDir := "/tmp/logs/"
	metricsLog := "test-metrics.log"
	perm := 0755
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no any file in baseDir,create new file",
			fields: fields{
				baseDir:           baseDir,
				baseFilename:      metricsLog,
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: 0,
				latestOpSec:       0,
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				time: uint64(time.Date(2023, 12, 15, 0, 0, 0, 0, time.Local).Unix() * 1000),
				before: func() {
					_ = os.RemoveAll(baseDir)
					_ = os.MkdirAll(baseDir, os.FileMode(perm))
				},
				clear: func() {
					_ = os.RemoveAll(baseDir)
				},
			},
			want:    baseDir + metricsLog,
			wantErr: assert.NoError,
		},
		{
			name: "have today date file in baseDir,rotate it and create new file",
			fields: fields{
				baseDir:           baseDir,
				baseFilename:      metricsLog,
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       time.Date(2023, 12, 15, 10, 0, 0, 0, time.Local).Unix(),
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				time: uint64(time.Date(2023, 12, 15, 10, 0, 1, 0, time.Local).Unix() * 1000),
				before: func() {
					_ = os.RemoveAll(baseDir)
					_ = os.MkdirAll(baseDir, os.FileMode(perm))
				},
				file: func() {
					_, _ = os.Create(baseDir + metricsLog)
				},
				clear: func() {
					_ = os.RemoveAll(baseDir)
				},
			},
			want:    baseDir + metricsLog,
			wantErr: assert.NoError,
		},
		{
			name: "have yesterday date file in baseDir,rotate it to yesterday date file and create new today date file",
			fields: fields{
				baseDir:           baseDir,
				baseFilename:      metricsLog,
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       time.Date(2023, 12, 14, 23, 59, 59, 0, time.Local).Unix(), //昨日
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				time: uint64(time.Date(2023, 12, 15, 0, 0, 0, 0, time.Local).Unix() * 1000),
				before: func() {
					_ = os.RemoveAll(baseDir)
					_ = os.MkdirAll(baseDir, os.FileMode(perm))
				},
				file: func() {
					yesterday := time.Date(2023, 12, 14, 23, 59, 59, 0, time.Local).Unix() * 1000
					dateStr := util.FormatDate(uint64(yesterday))
					_, _ = os.Create(baseDir + metricsLog)
					_, _ = os.Create(baseDir + metricsLog + "." + dateStr + ".1")
				},
				clear: func() {
					_ = os.RemoveAll(baseDir)
				},
			},
			want:    baseDir + metricsLog,
			wantErr: assert.NoError,
		},
		{
			name: "have yesterday archive file and today date file in baseDir,rotate it and create new file",
			fields: fields{
				baseDir:           baseDir,
				baseFilename:      metricsLog,
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       time.Date(2023, 12, 15, 10, 0, 0, 0, time.Local).Unix(),
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				time: uint64(time.Date(2023, 12, 15, 10, 0, 1, 0, time.Local).Unix() * 1000),
				before: func() {
					_ = os.RemoveAll(baseDir)
					_ = os.MkdirAll(baseDir, os.FileMode(perm))
				},
				file: func() {
					yesterday := time.Now().Add(-24 * time.Hour).Unix() * 1000
					dateStr := util.FormatDate(uint64(yesterday))
					_, _ = os.Create(baseDir + metricsLog)
					_, _ = os.Create(baseDir + metricsLog + "." + dateStr + ".1")
				},
				clear: func() {
					_ = os.RemoveAll(baseDir)
				},
			},
			want:    baseDir + metricsLog,
			wantErr: assert.NoError,
		},
		{
			name: "have yesterday archive file and today archive file and today date file in baseDir,rotate it and create new file",
			fields: fields{
				baseDir:           baseDir,
				baseFilename:      metricsLog,
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       time.Date(2023, 12, 15, 10, 0, 0, 0, time.Local).Unix(),
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				time: uint64(time.Date(2023, 12, 15, 10, 1, 1, 0, time.Local).Unix() * 1000),
				before: func() {
					_ = os.RemoveAll(baseDir)
					_ = os.MkdirAll(baseDir, os.FileMode(perm))
				},
				file: func() {
					yesterday := time.Now().Add(-24 * time.Hour).Unix() * 1000
					dateStr := util.FormatDate(uint64(yesterday))
					todayDateStr := util.FormatDate(uint64(time.Now().Unix() * 1000))
					_, _ = os.Create(baseDir + metricsLog)
					_, _ = os.Create(baseDir + metricsLog + "." + todayDateStr + ".1")
					_, _ = os.Create(baseDir + metricsLog + "." + dateStr + ".1")
				},
				clear: func() {
					_ = os.RemoveAll(baseDir)
				},
			},
			want:    baseDir + metricsLog,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultMetricLogWriter{
				baseDir:           tt.fields.baseDir,
				baseFilename:      tt.fields.baseFilename,
				maxSingleSize:     tt.fields.maxSingleSize,
				maxFileAmount:     tt.fields.maxFileAmount,
				timezoneOffsetSec: tt.fields.timezoneOffsetSec,
				latestOpSec:       tt.fields.latestOpSec,
				curMetricFile:     tt.fields.curMetricFile,
				curMetricIdxFile:  tt.fields.curMetricIdxFile,
				metricOut:         tt.fields.metricOut,
				idxOut:            tt.fields.idxOut,
				mux:               tt.fields.mux,
			}
			if tt.args.before != nil {
				tt.args.before()
			}
			if tt.args.file != nil {
				tt.args.file()
			}
			newFileName, err := d.rotateWithDateAndNewFileName(tt.args.time)
			if !tt.wantErr(t, err, fmt.Sprintf("rotateWithDateAndNewFileName(%v)", tt.args.time)) {
				return
			}
			_, _ = os.Create(newFileName)
			if tt.args.clear != nil {
				tt.args.clear()
			}
			assert.Equalf(t, tt.want, newFileName, "rotateWithDateAndNewFileName(%v)", tt.args.time)
		})
	}
}

func TestDefaultMetricLogWriter_isNewDay(t *testing.T) {
	_, offset := util.Now().Zone()
	fmt.Println("offset:", offset)
	type fields struct {
		baseDir           string
		baseFilename      string
		maxSingleSize     uint64
		maxFileAmount     uint32
		timezoneOffsetSec int64
		latestOpSec       int64
		curMetricFile     *os.File
		curMetricIdxFile  *os.File
		metricOut         *bufio.Writer
		idxOut            *bufio.Writer
		mux               *sync.RWMutex
	}
	type args struct {
		lastSec int64
		sec     int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "isNewDay true",
			fields: fields{
				baseDir:           "/tmp/logs/",
				baseFilename:      "test-metrics.log",
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       0,
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},

			args: args{
				lastSec: time.Date(2023, 12, 14, 23, 59, 59, 0, time.Local).Unix(),
				sec:     time.Date(2023, 12, 15, 0, 0, 0, 0, time.Local).Unix(),
			},
			want: true,
		},
		{
			name: "isNewDay false",
			fields: fields{
				baseDir:           "/tmp/logs/",
				baseFilename:      "test-metrics.log",
				maxSingleSize:     1024,
				maxFileAmount:     10,
				timezoneOffsetSec: int64(offset),
				latestOpSec:       0,
				curMetricFile:     nil,
				curMetricIdxFile:  nil,
				metricOut:         nil,
				idxOut:            nil,
				mux:               nil,
			},
			args: args{
				lastSec: time.Date(2023, 12, 15, 0, 0, 0, 0, time.Local).Unix(),
				sec:     time.Date(2023, 12, 15, 23, 59, 59, 0, time.Local).Unix(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultMetricLogWriter{
				baseDir:           tt.fields.baseDir,
				baseFilename:      tt.fields.baseFilename,
				maxSingleSize:     tt.fields.maxSingleSize,
				maxFileAmount:     tt.fields.maxFileAmount,
				timezoneOffsetSec: tt.fields.timezoneOffsetSec,
				latestOpSec:       tt.fields.latestOpSec,
				curMetricFile:     tt.fields.curMetricFile,
				curMetricIdxFile:  tt.fields.curMetricIdxFile,
				metricOut:         tt.fields.metricOut,
				idxOut:            tt.fields.idxOut,
				mux:               tt.fields.mux,
			}
			assert.Equalf(t, tt.want, d.isNewDay(tt.args.lastSec, tt.args.sec), "isNewDay(%v, %v)", tt.args.lastSec, tt.args.sec)
		})
	}
}
