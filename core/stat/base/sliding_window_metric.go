package base

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
	"math"
	"sync/atomic"
)

// SlidingWindowMetric represents the sliding window metric wrapper.
// It does not store any data and is the wrapper of BucketLeapArray to adapt to different internal window
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

// Get the start time range of the bucket for the provided time.
// The actual time span is: [start, end + in.bucketTimeLength)
func (m *SlidingWindowMetric) getBucketStartRange(timeMs uint64) (start, end uint64) {
	curBucketStartTime := calculateStartTime(timeMs, m.real.WindowLengthInMs())
	end = curBucketStartTime
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
			logger.Error("Illegal state: current bucket value is nil when summing count")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logger.Errorf("Fail to cast data value(%+v) to MetricBucket type", mb)
			continue
		}
		ret += counter.Get(event)
	}
	return ret
}

func (m *SlidingWindowMetric) GetSum(event base.MetricEvent) int64 {
	return m.getSumWithTime(util.CurrentTimeMillis(), event)
}

func (m *SlidingWindowMetric) getSumWithTime(now uint64, event base.MetricEvent) int64 {
	start, end := m.getBucketStartRange(now)
	satisfiedBuckets := m.real.ValuesConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	return m.count(event, satisfiedBuckets)
}

func (m *SlidingWindowMetric) GetAvg(event base.MetricEvent) float64 {
	return m.getAvgWithTime(util.CurrentTimeMillis(), event)
}

func (m *SlidingWindowMetric) getAvgWithTime(now uint64, event base.MetricEvent) float64 {
	return float64(m.getSumWithTime(now, event)) / m.getIntervalInSecond()
}

func (m *SlidingWindowMetric) MinRT() int64 {
	now := util.CurrentTimeMillis()
	start, end := m.getBucketStartRange(now)
	satisfiedBuckets := m.real.ValuesConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	minRt := int64(math.MaxInt64)
	for _, w := range satisfiedBuckets {
		mb := w.value.Load()
		if mb == nil {
			logger.Error("Illegal state: current bucket value is nil when calculating min")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logger.Errorf("Fail to cast data value(%+v) to MetricBucket type", mb)
			continue
		}
		rt := counter.Get(base.MetricEventRt)
		if rt < minRt {
			minRt = rt
		}
	}
	return minRt
}

func (m *SlidingWindowMetric) AvgRT() float64 {
	return float64(m.GetSum(base.MetricEventRt)) / float64(m.GetSum(base.MetricEventComplete))
}

// SecondMetricsOnCondition aggregates metric items by second on condition that
// the startTime of the statistic buckets satisfies the time predicate.
func (m *SlidingWindowMetric) SecondMetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	ws := m.real.ValuesConditional(util.CurrentTimeMillis(), predicate)

	// Aggregate second-level MetricItem (only for stable metrics)
	wm := make(map[uint64][]*windowWrap)
	for _, w := range ws {
		bucketStart := atomic.LoadUint64(&w.bucketStart)
		secStart := bucketStart - bucketStart%1000
		if arr, hasData := wm[secStart]; hasData {
			wm[secStart] = append(arr, w)
		} else {
			wm[secStart] = []*windowWrap{w}
		}
	}
	items := make([]*base.MetricItem, 0)
	for ts, values := range wm {
		if len(values) == 0 {
			continue
		}
		if item := m.metricItemFromBuckets(ts, values); item != nil {
			items = append(items, item)
		}
	}
	return items
}

// metricItemFromBuckets aggregates multiple bucket wrappers (based on the same startTime in second)
// to the single MetricItem.
func (m *SlidingWindowMetric) metricItemFromBuckets(ts uint64, ws []*windowWrap) *base.MetricItem {
	item := &base.MetricItem{Timestamp: ts}
	var allRt int64 = 0
	for _, w := range ws {
		mi := w.value.Load()
		if mi == nil {
			logger.Error("Get nil bucket when generating MetricItem from buckets")
			return nil
		}
		mb, ok := mi.(*MetricBucket)
		if !ok {
			logger.Errorf("Failed to cast to MetricBucket type, bucket startTime: %d", w.bucketStart)
			return nil
		}
		item.PassQps += uint64(mb.Get(base.MetricEventPass))
		item.BlockQps += uint64(mb.Get(base.MetricEventBlock))
		item.ErrorQps += uint64(mb.Get(base.MetricEventError))
		item.CompleteQps += uint64(mb.Get(base.MetricEventComplete))
		allRt += mb.Get(base.MetricEventRt)
	}
	if item.CompleteQps > 0 {
		item.AvgRt = uint64(allRt) / item.CompleteQps
	} else {
		item.AvgRt = uint64(allRt)
	}
	return item
}

func (m *SlidingWindowMetric) metricItemFromBucket(w *windowWrap) *base.MetricItem {
	mi := w.value.Load()
	if mi == nil {
		logger.Error("Get nil bucket when generating MetricItem from buckets")
		return nil
	}
	mb, ok := mi.(*MetricBucket)
	if !ok {
		logger.Errorf("Fail to cast data value to MetricBucket type, bucket startTime: %d", w.bucketStart)
		return nil
	}
	completeQps := mb.Get(base.MetricEventComplete)
	item := &base.MetricItem{
		PassQps:     uint64(mb.Get(base.MetricEventPass)),
		BlockQps:    uint64(mb.Get(base.MetricEventBlock)),
		ErrorQps:    uint64(mb.Get(base.MetricEventError)),
		CompleteQps: uint64(completeQps),
		Timestamp:   w.bucketStart,
	}
	if completeQps > 0 {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt) / completeQps)
	} else {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt))
	}
	return item
}
