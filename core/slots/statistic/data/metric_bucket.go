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
)
const metricEventNum = 5

/**
MetricBucket store the metric statistic of each event
(MetricEventPass、MetricEventBlock、MetricEventError、MetricEventSuccess、MetricEventRt)
*/
type MetricBucket struct {
	counter []uint64
	minRt   uint64
}

func (mb *MetricBucket) metricEvents() []MetricEventType {
	met := make([]MetricEventType, 0, metricEventNum)
	met = append(met, MetricEventPass)
	met = append(met, MetricEventBlock)
	met = append(met, MetricEventError)
	met = append(met, MetricEventSuccess)
	met = append(met, MetricEventRt)
	return met
}

func newEmptyMetricBucket() MetricBucket {
	mb := MetricBucket{
		counter: make([]uint64, metricEventNum, metricEventNum),
		minRt:   math.MaxUint64,
	}
	return mb
}

func (mb *MetricBucket) Add(event MetricEventType, count uint64) {
	switch event {
	case MetricEventPass:
		atomic.AddUint64(&mb.counter[0], count)
	case MetricEventBlock:
		atomic.AddUint64(&mb.counter[1], count)
	case MetricEventError:
		atomic.AddUint64(&mb.counter[2], count)
	case MetricEventSuccess:
		atomic.AddUint64(&mb.counter[3], count)
	case MetricEventRt:
		atomic.AddUint64(&mb.counter[4], count)
	default:
		panic("unknown metric event type, " + string(event))
	}
}

func (mb *MetricBucket) Get(event MetricEventType) uint64 {
	switch event {
	case MetricEventPass:
		return atomic.LoadUint64(&mb.counter[0])
	case MetricEventBlock:
		return atomic.LoadUint64(&mb.counter[1])
	case MetricEventError:
		return atomic.LoadUint64(&mb.counter[2])
	case MetricEventSuccess:
		return atomic.LoadUint64(&mb.counter[3])
	case MetricEventRt:
		return atomic.LoadUint64(&mb.counter[4])
	default:
		panic("unknown metric event type, " + string(event))
	}
}

func (mb *MetricBucket) AddRt(rt uint64) {
	mb.Add(MetricEventRt, rt)
	if rt < mb.minRt {
		mb.minRt = rt
	}
}

func (mb *MetricBucket) Reset() {
	for i := 0; i < metricEventNum; i++ {
		atomic.StoreUint64(&mb.counter[i], 0)
	}
	mb.minRt = math.MaxUint64
}
