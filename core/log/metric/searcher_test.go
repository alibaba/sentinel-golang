package metric

import (
	"reflect"
	"sync"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
)

func TestDefaultMetricSearcher_FindByTimeAndResource(t *testing.T) {
	reader := newDefaultMetricLogReader()
	cachedPos := &filePosition{}
	mux := new(sync.Mutex)

	type fields struct {
		reader       MetricLogReader
		baseDir      string
		baseFilename string
		cachedPos    *filePosition
		mux          *sync.Mutex
	}
	type args struct {
		beginTimeMs uint64
		endTimeMs   uint64
		resource    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*base.MetricItem
		wantErr bool
	}{
		{
			"test1",
			fields{
				reader:       reader,
				baseDir:      ".../.../../",
				baseFilename: ".",
				cachedPos:    cachedPos,
				mux:          mux,
			},
			args{
				beginTimeMs: 19999999,
				endTimeMs:   19999999,
				resource:    "res",
			},
			nil,
			true,
		}, {
			"test2",
			fields{
				reader:       reader,
				baseDir:      "../../../tests/testdata/metric",
				baseFilename: "app1-metrics.log",
				cachedPos:    cachedPos,
				mux:          mux,
			},
			args{
				beginTimeMs: 19999999,
				endTimeMs:   19999999,
				resource:    "res",
			},
			[]*base.MetricItem{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DefaultMetricSearcher{
				reader:       tt.fields.reader,
				baseDir:      tt.fields.baseDir,
				baseFilename: tt.fields.baseFilename,
				cachedPos:    tt.fields.cachedPos,
				mux:          tt.fields.mux,
			}
			got, err := s.FindByTimeAndResource(tt.args.beginTimeMs, tt.args.endTimeMs, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultMetricSearcher.FindByTimeAndResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultMetricSearcher.FindByTimeAndResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultMetricSearcher_FindFromTimeWithMaxLines(t *testing.T) {
	reader := newDefaultMetricLogReader()
	cachedPos := &filePosition{}
	mux := new(sync.Mutex)

	type fields struct {
		reader       MetricLogReader
		baseDir      string
		baseFilename string
		cachedPos    *filePosition
		mux          *sync.Mutex
	}
	type args struct {
		beginTimeMs uint64
		maxLines    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*base.MetricItem
		wantErr bool
	}{
		{
			"test1",
			fields{
				reader:       reader,
				baseDir:      ".../.../../",
				baseFilename: ".",
				cachedPos:    cachedPos,
				mux:          mux,
			},
			args{
				beginTimeMs: 19999999,
				maxLines:    19999999,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DefaultMetricSearcher{
				reader:       tt.fields.reader,
				baseDir:      tt.fields.baseDir,
				baseFilename: tt.fields.baseFilename,
				cachedPos:    tt.fields.cachedPos,
				mux:          tt.fields.mux,
			}
			got, err := s.FindFromTimeWithMaxLines(tt.args.beginTimeMs, tt.args.maxLines)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultMetricSearcher.FindFromTimeWithMaxLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultMetricSearcher.FindFromTimeWithMaxLines() = %v, want %v", got, tt.want)
			}
		})
	}
}
