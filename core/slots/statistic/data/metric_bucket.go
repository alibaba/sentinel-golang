package data

import (
	"math"
	"sync/atomic"
)

type MetricEventType int8

const (
	MetricEventPass MetricEventType = iota
	MetricEventBlock
	MetricEventSuccess
	MetricEventError
	MetricEventRt
	// hack for getting length of enum
	metricEventNum
)

/**
MetricBucket store the metric statistic of each event
(MetricEventPass、MetricEventBlock、MetricEventError、MetricEventSuccess、MetricEventRt)
*/
type MetricBucket struct {
	// value of statistic
	counters [metricEventNum]uint64
	minRt    uint64
}

func newMetricBucket() MetricBucket {
	mb := MetricBucket{
		minRt: math.MaxUint64,
	}
	return mb
}

func (mb *MetricBucket) Add(event MetricEventType, count uint64) {
	if event > metricEventNum {
		panic("event is bigger then metricEventNum")
	}
	atomic.AddUint64(&mb.counters[event], count)
}

func (mb *MetricBucket) Get(event MetricEventType) uint64 {
	if event > metricEventNum {
		panic("event is bigger then metricEventNum")
	}
	return mb.counters[event]
}

func (mb *MetricBucket) MinRt() uint64 {
	return mb.minRt
}

func (mb *MetricBucket) Reset() {
	for i := 0; i < int(metricEventNum); i++ {
		atomic.StoreUint64(&mb.counters[i], 0)
	}
	atomic.StoreUint64(&mb.minRt, math.MaxUint64)
}

func (mb *MetricBucket) AddPass(n uint64) {
	mb.Add(MetricEventPass, n)
}

func (mb *MetricBucket) Pass() uint64 {
	return mb.Get(MetricEventPass)
}

func (mb *MetricBucket) AddBlock(n uint64) {
	mb.Add(MetricEventBlock, n)
}

func (mb *MetricBucket) Block() uint64 {
	return mb.Get(MetricEventBlock)
}

func (mb *MetricBucket) AddSuccess(n uint64) {
	mb.Add(MetricEventSuccess, n)
}

func (mb *MetricBucket) Success() uint64 {
	return mb.Get(MetricEventSuccess)
}

func (mb *MetricBucket) AddError(n uint64) {
	mb.Add(MetricEventError, n)
}

func (mb *MetricBucket) Error() uint64 {
	return mb.Get(MetricEventError)
}

func (mb *MetricBucket) AddRt(rt uint64) {
	mb.Add(MetricEventRt, rt)
	// Not thread-safe, but it's okay.
	if rt < mb.minRt {
		mb.minRt = rt
	}
}

func (mb *MetricBucket) Rt() uint64 {
	return mb.Get(MetricEventRt)
}
