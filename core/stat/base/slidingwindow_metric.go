package base

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
	"math"
)

// The basic metric struct,
// SlidingWindowMetric doesn't store any data and is the wrapper of BucketLeapArray to adapt to different internal window
// SlidingWindowMetric is used for SentinelRules and BucketLeapArray is used for monitor
// BucketLeapArray is per resource, and SlidingWindowMetric support only read operation.
type SlidingWindowMetric struct {
	windowLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	real             *BucketLeapArray
}

// It must pass the parameter point to the real storage entity
func NewSlidingWindowMetric(sampleCount, intervalInMs uint32, real *BucketLeapArray) *SlidingWindowMetric {
	if real == nil || intervalInMs <= 0 || sampleCount <= 0 {
		panic(fmt.Sprintf("Illegal parameters,intervalInMs=%d,sampleCount=%d,real=%+v.", intervalInMs, sampleCount, real))
	}

	if intervalInMs%sampleCount != 0 {
		panic(fmt.Sprintf("Invalid parameters, intervalInMs is %d, sampleCount is %d.", intervalInMs, sampleCount))
	}
	winLengthInMs := intervalInMs / sampleCount

	parentIntervalInMs := real.IntervalInMs()
	parentWindowLengthInMs := real.WindowLengthInMs()

	// winLengthInMs of BucketLeapArray must be divisible by winLengthInMs of SlidingWindowMetric
	// for example: winLengthInMs of BucketLeapArray is 500ms, and winLengthInMs of SlidingWindowMetric is 2000ms
	// for example: winLengthInMs of BucketLeapArray is 500ms, and winLengthInMs of SlidingWindowMetric is 500ms
	if winLengthInMs%parentWindowLengthInMs != 0 {
		panic(fmt.Sprintf("BucketLeapArray's WindowLengthInMs(%d) is not divisible by SlidingWindowMetric's WindowLengthInMs(%d).", parentWindowLengthInMs, winLengthInMs))
	}

	if intervalInMs > parentIntervalInMs {
		// todo if SlidingWindowMetric's intervalInMs is greater than BucketLeapArray.
		panic(fmt.Sprintf("The interval(%d) of SlidingWindowMetric is greater than parent BucketLeapArray(%d).", intervalInMs, parentIntervalInMs))
	}

	// 10 * 1000 ms == parent
	if parentIntervalInMs%intervalInMs != 0 {
		panic(fmt.Sprintf("SlidingWindowMetric's intervalInMs(%d) is not divisible by real BucketLeapArray's intervalInMs(%d).", intervalInMs, parentIntervalInMs))
	}

	return &SlidingWindowMetric{
		windowLengthInMs: winLengthInMs,
		sampleCount:      sampleCount,
		intervalInMs:     intervalInMs,
		real:             real,
	}
}

// Get the [start time, end time) of SlidingWindowMetric for now.
func (m *SlidingWindowMetric) getTimeInterval(now uint64) (start, end uint64) {
	curWindowWrapStartTime := calculateStartTime(now, m.real.WindowLengthInMs())
	end = curWindowWrapStartTime
	start = end - uint64(m.intervalInMs) + uint64(m.real.WindowLengthInMs())
	return
}

func (m *SlidingWindowMetric) getIntervalInSecond() float64 {
	return float64(m.intervalInMs) / 1000.0
}

func (m *SlidingWindowMetric) count(event base.MetricEvent, values []*windowWrap) int64 {
	ret := int64(0)
	for _, ww := range values {
		mb := ww.value.Load()
		if mb == nil {
			logger.Error("Current window's value is nil.")
			continue
		}
		counter, ok := mb.(*metricBucket)
		if !ok {
			logger.Errorf("Fail to assert windowWrap's value(%+v) to metricBucket.", mb)
			continue
		}
		ret += counter.get(event)
	}
	return ret
}

func (m *SlidingWindowMetric) GetQPS(event base.MetricEvent) float64 {
	now := util.CurrentTimeMillis()
	return m.GetQPSWithTime(now, event)
}

func (m *SlidingWindowMetric) GetQPSWithTime(now uint64, event base.MetricEvent) float64 {
	start, end := m.getTimeInterval(now)
	satisfiedBuckets := m.real.ValuesWithConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	cnt := m.count(event, satisfiedBuckets)
	return float64(cnt) / m.getIntervalInSecond()
}

func (m *SlidingWindowMetric) GetSum(event base.MetricEvent) int64 {
	now := util.CurrentTimeMillis()
	return m.GetSumWithTime(now, event)
}

func (m *SlidingWindowMetric) GetSumWithTime(now uint64, event base.MetricEvent) int64 {
	start, end := m.getTimeInterval(now)
	satisfiedBuckets := m.real.ValuesWithConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	return m.count(event, satisfiedBuckets)
}

func (m *SlidingWindowMetric) MinRT() int64 {
	now := util.CurrentTimeMillis()
	start, end := m.getTimeInterval(now)
	satisfiedBuckets := m.real.ValuesWithConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end

	})
	minRt := int64(math.MaxInt64)
	for _, w := range satisfiedBuckets {
		mb := w.value.Load()
		if mb == nil {
			logger.Error("Current window's value is nil.")
			continue
		}
		counter, ok := mb.(*metricBucket)
		if !ok {
			logger.Errorf("Fail to assert windowWrap's value(%+v) to metricBucket.", mb)
			continue
		}
		rt := counter.get(base.MetricEventRt)
		if rt < minRt {
			minRt = rt
		}
	}
	return minRt
}

func (m *SlidingWindowMetric) AvgRT() float64 {
	return float64(m.GetSum(base.MetricEventRt)) / float64(m.GetSum(base.MetricEventComplete))
}
