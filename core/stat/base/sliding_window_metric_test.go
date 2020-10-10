package base

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
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
			m, err := NewSlidingWindowMetric(tt.args.sampleCount, tt.args.intervalInMs, NewBucketLeapArray(tt.args.realSampleCount, tt.args.realIntervalInMs))
			assert.True(t, err == nil)

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
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	got, err = NewSlidingWindowMetric(0, 0, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, got == nil && err != nil)
	got, err = NewSlidingWindowMetric(4, 2001, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, got == nil && err != nil)
	got, err = NewSlidingWindowMetric(2, 2002, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, got == nil && err != nil)
	got, err = NewSlidingWindowMetric(4, 200000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, got == nil && err != nil)
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
			m, _ := NewSlidingWindowMetric(tt.fields.sampleCount, tt.fields.intervalInMs, tt.fields.real)
			if got := m.getSumWithTime(tt.args.now, tt.args.event); got != tt.want {
				t.Errorf("SlidingWindowMetric.getSumWithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSum(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	passSum := got.GetSum(base.MetricEventPass)
	assert.True(t, passSum == 0)
}

func TestGetQPS(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	qps := got.GetQPS(base.MetricEventPass)
	assert.True(t, qps == 0)
}

func TestGetPreviousQPS(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	previousQPS := got.GetPreviousQPS(base.MetricEventPass)
	assert.True(t, previousQPS == 0)
}

func TestGetQPSWithTime(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	qps := got.getQPSWithTime(util.CurrentTimeMillis(), base.MetricEventPass)
	assert.True(t, qps == 0)
}

func TestGetMaxOfSingleBucket(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	got.real.AddCount(base.MetricEventPass, 100)
	max := got.GetMaxOfSingleBucket(base.MetricEventPass)
	assert.True(t, max == 100)
}

func TestMinRT(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	minRt := got.MinRT()
	assert.True(t, minRt == float64(base.DefaultStatisticMaxRt))
}

func TestAvgRT(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	got.real.AddCount(base.MetricEventRt, 100)
	got.real.AddCount(base.MetricEventComplete, 100)
	avgRT := got.AvgRT()
	assert.True(t, avgRT == 1)
}

func TestMetricItemFromBuckets(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	got.real.AddCount(base.MetricEventPass, 100)
	item := got.metricItemFromBuckets(util.CurrentTimeMillis(), got.real.data.array.data)
	assert.True(t, item.PassQps == 100)
}

func TestMetricItemFromBucket(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	got.real.AddCount(base.MetricEventPass, 100)
	currentBucket, err := got.real.data.CurrentBucket(&leapArrayMock{})
	assert.Nil(t, err)
	item := got.metricItemFromBucket(currentBucket)
	assert.True(t, item.PassQps == 100)
}

func TestSecondMetricsOnCondition(t *testing.T) {
	got, err := NewSlidingWindowMetric(4, 2000, NewBucketLeapArray(SampleCount, IntervalInMs))
	assert.True(t, err == nil && got != nil)
	start, end := got.getBucketStartRange(util.CurrentTimeMillis())
	items := got.SecondMetricsOnCondition(func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	assert.True(t, len(items) == 1)
}
