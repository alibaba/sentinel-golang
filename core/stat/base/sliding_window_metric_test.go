package base

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
)

func TestSlidingWindowMetric_getBucketStartRange(t *testing.T) {
	type args struct {
		sampleCount      uint32
		intervalInMs     uint32
		realSampleCount  uint32
		realIntervalInMs uint32
		now              uint64
	}
	tests := []struct {
		name      string
		args      args
		wantStart uint64
		wantEnd   uint64
	}{
		{
			name: "TestSlidingWindowMetric_getBucketStartRange-1",
			args: args{
				sampleCount:      4,
				intervalInMs:     2000,
				realSampleCount:  20,
				realIntervalInMs: 10000,
				// array start time:1578416550000
				// bucket start time:1578416556500
				now: 1578416556900, //
			},
			wantStart: 1578416555000,
			wantEnd:   1578416556500,
		},
		{
			name: "TestSlidingWindowMetric_getBucketStartRange-2",
			args: args{
				sampleCount:      2,
				intervalInMs:     1000,
				realSampleCount:  20,
				realIntervalInMs: 10000,
				// array start time:1578416550000
				// bucket start time:1578416556500
				now: 1578416556900, //
			},
			wantStart: 1578416556000,
			wantEnd:   1578416556500,
		},
		{
			name: "TestSlidingWindowMetric_getBucketStartRange-3",
			args: args{
				sampleCount:      1,
				intervalInMs:     2000,
				realSampleCount:  10,
				realIntervalInMs: 10000,
				// array start time:1578416550000
				// bucket start time:1578416556500
				now: 1578416556900, //
			},
			wantStart: 1578416555000,
			wantEnd:   1578416556000,
		},
		{
			name: "TestSlidingWindowMetric_getBucketStartRange-4",
			args: args{
				sampleCount:      1,
				intervalInMs:     10000,
				realSampleCount:  10,
				realIntervalInMs: 20000,
				// array start time:1578416550000
				// bucket start time:1578416556500
				now: 1578416556900, //
			},
			wantStart: 1578416548000,
			wantEnd:   1578416556000,
		},
		{
			name: "TestSlidingWindowMetric_getBucketStartRange-5",
			args: args{
				sampleCount:      2,
				intervalInMs:     1000,
				realSampleCount:  20,
				realIntervalInMs: 10000,
				// array start time:1578416550000
				// bucket start time:1578416556500
				now: 1578416556500, //
			},
			wantStart: 1578416556000,
			wantEnd:   1578416556500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSlidingWindowMetric(tt.args.sampleCount, tt.args.intervalInMs, NewBucketLeapArray(tt.args.realSampleCount, tt.args.realIntervalInMs))
			gotStart, gotEnd := m.getBucketStartRange(tt.args.now)
			if gotStart != tt.wantStart {
				t.Errorf("SlidingWindowMetric.getBucketStartRange() gotStart = %v, want %v", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("SlidingWindowMetric.getBucketStartRange() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

func Test_NewSlidingWindowMetric(t *testing.T) {
	type args struct {
		sampleCount  uint32
		intervalInMs uint32
		real         *BucketLeapArray
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test_NewSlidingWindowMetric-1",
			args: args{
				intervalInMs: 2000,
				sampleCount:  4,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			want: "",
		},
		{
			name: "Test_NewSlidingWindowMetric-2",
			args: args{
				intervalInMs: 0,
				sampleCount:  0,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			want: "Illegal parameters,intervalInMs=0,sampleCount=0,real=",
		},
		{
			name: "Test_NewSlidingWindowMetric-3",
			args: args{
				intervalInMs: 2001,
				sampleCount:  4,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			want: "Invalid parameters, intervalInMs is 2001, sampleCount is 4.",
		},
		{
			name: "Test_NewSlidingWindowMetric-4",
			args: args{
				intervalInMs: 2002,
				sampleCount:  2,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			want: "BucketLeapArray's BucketLengthInMs(500) is not divisible by SlidingWindowMetric's BucketLengthInMs(1001).",
		},
		{
			name: "Test_NewSlidingWindowMetric-5",
			args: args{
				intervalInMs: 200000,
				sampleCount:  4,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			want: "The interval(200000) of SlidingWindowMetric is greater than parent BucketLeapArray(10000).",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					errContent, ok := err.(string)
					if !ok {
						t.Errorf("Fail to assert err, except string, in fact:%+v", reflect.TypeOf(err))
					}
					if !strings.Contains(errContent, tt.want) {
						t.Errorf("Failed, except [%s],in fact:[%s]", tt.want, errContent)
					}
				}
			}()
			got := NewSlidingWindowMetric(tt.args.sampleCount, tt.args.intervalInMs, tt.args.real)
			if got == nil || "" != tt.want {
				t.Errorf("NewSlidingWindowMetric() = %v", got)
			}
		})
	}
}

func TestSlidingWindowMetric_GetIntervalSumWithTime(t *testing.T) {
	type fields struct {
		sampleCount  uint32
		intervalInMs uint32
		real         *BucketLeapArray
	}
	type args struct {
		event base.MetricEvent
		now   uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "",
			fields: fields{
				sampleCount:  2,
				intervalInMs: 2000,
				real:         NewBucketLeapArray(SampleCount, IntervalInMs),
			},
			args: args{
				event: base.MetricEventPass,
				now:   1678416556599,
			},
			want: 2000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 500; i++ {
				tt.fields.real.addCountWithTime(tt.args.now, tt.args.event, 1)
			}
			for i := 0; i < int(tt.fields.intervalInMs); i++ {
				tt.fields.real.addCountWithTime(tt.args.now-100-uint64(i), tt.args.event, 1)
			}
			m := NewSlidingWindowMetric(tt.fields.sampleCount, tt.fields.intervalInMs, tt.fields.real)
			if got := m.getSumWithTime(tt.args.now, tt.args.event); got != tt.want {
				t.Errorf("SlidingWindowMetric.getSumWithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
