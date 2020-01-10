package base

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/base"
	"sync/atomic"
)

// metricBucket is the storage entity to record metrics
// Event type contains (MetricEventPass、MetricEventBlock、MetricEventError、MetricEventComplete、MetricEventRt)
// The statistic of metricBucket must be concurrent safe.
// In order to save memory, metricBucket only counts
// Each metricBucket need 8*5 = 40 bytes
type metricBucket struct {
	// value of statistic
	counter [base.MetricEventTotal]int64
	minRt   int64
}

func newMetricBucket() *metricBucket {
	mb := &metricBucket{
		minRt: base.DefaultStatisticMaxRt,
	}
	return mb
}

func (mb *metricBucket) add(event base.MetricEvent, count int64) {
	if event > base.MetricEventTotal || event < 0 {
		panic(fmt.Sprintf("Event %v is unknown.", event))
	}
	atomic.AddInt64(&mb.counter[event], count)
}

func (mb *metricBucket) get(event base.MetricEvent) int64 {
	if event > base.MetricEventTotal || event < 0 {
		panic(fmt.Sprintf("Event %v is unknown.", event))
	}
	return mb.counter[event]
}

func (mb *metricBucket) reset() {
	for i := 0; i < int(base.MetricEventTotal); i++ {
		atomic.StoreInt64(&mb.counter[i], 0)
	}
	atomic.StoreInt64(&mb.minRt, base.DefaultStatisticMaxRt)
}

func (mb *metricBucket) addPass(n int64) {
	mb.add(base.MetricEventPass, n)
}

func (mb *metricBucket) getPass() int64 {
	return mb.get(base.MetricEventPass)
}

func (mb *metricBucket) addBlock(n int64) {
	mb.add(base.MetricEventBlock, n)
}

func (mb *metricBucket) getBlock() int64 {
	return mb.get(base.MetricEventBlock)
}

func (mb *metricBucket) addComplete(n int64) {
	mb.add(base.MetricEventComplete, n)
}

func (mb *metricBucket) getComplete() int64 {
	return mb.get(base.MetricEventComplete)
}

func (mb *metricBucket) addError(n int64) {
	mb.add(base.MetricEventError, n)
}

func (mb *metricBucket) getError() int64 {
	return mb.get(base.MetricEventError)
}

func (mb *metricBucket) addRt(rt int64) {
	mb.add(base.MetricEventRt, rt)
	if rt < atomic.LoadInt64(&mb.minRt) {
		atomic.StoreInt64(&mb.minRt, rt)
	}
}

func (mb *metricBucket) getRt() int64 {
	return mb.get(base.MetricEventRt)
}

func (mb *metricBucket) getMinRt() int64 {
	return atomic.LoadInt64(&mb.minRt)
}
