package base

import (
	"fmt"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

// SlidingWindowMetric represents the sliding window metric wrapper.
// It does not store any data and is the wrapper of BucketLeapArray to adapt to different internal bucket
// SlidingWindowMetric is used for SentinelRules and BucketLeapArray is used for monitor
// BucketLeapArray is per resource, and SlidingWindowMetric support only read operation.
type SlidingWindowMetric struct {
	bucketLengthInMs uint32
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
	bucketLengthInMs := intervalInMs / sampleCount

	parentIntervalInMs := real.IntervalInMs()
	parentBucketLengthInMs := real.BucketLengthInMs()

	// bucketLengthInMs of BucketLeapArray must be divisible by bucketLengthInMs of SlidingWindowMetric
	// for example: bucketLengthInMs of BucketLeapArray is 500ms, and bucketLengthInMs of SlidingWindowMetric is 2000ms
	// for example: bucketLengthInMs of BucketLeapArray is 500ms, and bucketLengthInMs of SlidingWindowMetric is 500ms
	if bucketLengthInMs%parentBucketLengthInMs != 0 {
		panic(fmt.Sprintf("BucketLeapArray's BucketLengthInMs(%d) is not divisible by SlidingWindowMetric's BucketLengthInMs(%d).", parentBucketLengthInMs, bucketLengthInMs))
	}

	if intervalInMs > parentIntervalInMs {
		// todo if SlidingWindowMetric's intervalInMs is greater than BucketLeapArray.
		panic(fmt.Sprintf("The interval(%d) of SlidingWindowMetric is greater than parent BucketLeapArray(%d).", intervalInMs, parentIntervalInMs))
	}

	return &SlidingWindowMetric{
		bucketLengthInMs: bucketLengthInMs,
		sampleCount:      sampleCount,
		intervalInMs:     intervalInMs,
		real:             real,
	}
}

// Get the start time range of the bucket for the provided time.
// The actual time span is: [start, end + in.bucketTimeLength)
func (m *SlidingWindowMetric) getBucketStartRange(timeMs uint64) (start, end uint64) {
	curBucketStartTime := calculateStartTime(timeMs, m.real.BucketLengthInMs())
	end = curBucketStartTime
	start = end - uint64(m.intervalInMs) + uint64(m.real.BucketLengthInMs())
	return
}

func (m *SlidingWindowMetric) getIntervalInSecond() float64 {
	return float64(m.intervalInMs) / 1000.0
}

func (m *SlidingWindowMetric) count(event base.MetricEvent, values []*BucketWrap) int64 {
	ret := int64(0)
	for _, ww := range values {
		mb := ww.Value.Load()
		if mb == nil {
			logging.Error("Illegal state: current bucket Value is nil when summing count")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Errorf("Fail to cast data Value(%+v) to MetricBucket type", mb)
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

func (m *SlidingWindowMetric) GetQPS(event base.MetricEvent) float64 {
	return m.getQPSWithTime(util.CurrentTimeMillis(), event)
}

func (m *SlidingWindowMetric) GetPreviousQPS(event base.MetricEvent) float64 {
	return m.getQPSWithTime(util.CurrentTimeMillis()-uint64(m.bucketLengthInMs), event)
}

func (m *SlidingWindowMetric) getQPSWithTime(now uint64, event base.MetricEvent) float64 {
	return float64(m.getSumWithTime(now, event)) / m.getIntervalInSecond()
}

func (m *SlidingWindowMetric) GetMaxOfSingleBucket(event base.MetricEvent) int64 {
	now := util.CurrentTimeMillis()
	start, end := m.getBucketStartRange(now)
	satisfiedBuckets := m.real.ValuesConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	var curMax int64 = 0
	for _, w := range satisfiedBuckets {
		mb := w.Value.Load()
		if mb == nil {
			logging.Error("Illegal state: current bucket Value is nil when GetMaxOfSingleBucket")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Errorf("Failed to cast data Value(%+v) to MetricBucket type", mb)
			continue
		}
		v := counter.Get(event)
		if v > curMax {
			curMax = v
		}
	}
	return curMax
}

func (m *SlidingWindowMetric) MinRT() float64 {
	now := util.CurrentTimeMillis()
	start, end := m.getBucketStartRange(now)
	satisfiedBuckets := m.real.ValuesConditional(now, func(ws uint64) bool {
		return ws >= start && ws <= end
	})
	minRt := base.DefaultStatisticMaxRt
	for _, w := range satisfiedBuckets {
		mb := w.Value.Load()
		if mb == nil {
			logging.Error("Illegal state: current bucket Value is nil when calculating minRT")
			continue
		}
		counter, ok := mb.(*MetricBucket)
		if !ok {
			logging.Errorf("Failed to cast data Value(%+v) to MetricBucket type", mb)
			continue
		}
		v := counter.MinRt()
		if v < minRt {
			minRt = v
		}
	}
	if minRt < 1 {
		minRt = 1
	}
	return float64(minRt)
}

func (m *SlidingWindowMetric) AvgRT() float64 {
	return float64(m.GetSum(base.MetricEventRt)) / float64(m.GetSum(base.MetricEventComplete))
}

// SecondMetricsOnCondition aggregates metric items by second on condition that
// the startTime of the statistic buckets satisfies the time predicate.
func (m *SlidingWindowMetric) SecondMetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	ws := m.real.ValuesConditional(util.CurrentTimeMillis(), predicate)

	// Aggregate second-level MetricItem (only for stable metrics)
	wm := make(map[uint64][]*BucketWrap)
	for _, w := range ws {
		bucketStart := atomic.LoadUint64(&w.BucketStart)
		secStart := bucketStart - bucketStart%1000
		if arr, hasData := wm[secStart]; hasData {
			wm[secStart] = append(arr, w)
		} else {
			wm[secStart] = []*BucketWrap{w}
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
func (m *SlidingWindowMetric) metricItemFromBuckets(ts uint64, ws []*BucketWrap) *base.MetricItem {
	item := &base.MetricItem{Timestamp: ts}
	var allRt int64 = 0
	for _, w := range ws {
		mi := w.Value.Load()
		if mi == nil {
			logging.Error("Get nil bucket when generating MetricItem from buckets")
			return nil
		}
		mb, ok := mi.(*MetricBucket)
		if !ok {
			logging.Errorf("Failed to cast to MetricBucket type, bucket startTime: %d", w.BucketStart)
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

func (m *SlidingWindowMetric) metricItemFromBucket(w *BucketWrap) *base.MetricItem {
	mi := w.Value.Load()
	if mi == nil {
		logging.Error("Get nil bucket when generating MetricItem from buckets")
		return nil
	}
	mb, ok := mi.(*MetricBucket)
	if !ok {
		logging.Errorf("Fail to cast data Value to MetricBucket type, bucket startTime: %d", w.BucketStart)
		return nil
	}
	completeQps := mb.Get(base.MetricEventComplete)
	item := &base.MetricItem{
		PassQps:     uint64(mb.Get(base.MetricEventPass)),
		BlockQps:    uint64(mb.Get(base.MetricEventBlock)),
		ErrorQps:    uint64(mb.Get(base.MetricEventError)),
		CompleteQps: uint64(completeQps),
		Timestamp:   w.BucketStart,
	}
	if completeQps > 0 {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt) / completeQps)
	} else {
		item.AvgRt = uint64(mb.Get(base.MetricEventRt))
	}
	return item
}
