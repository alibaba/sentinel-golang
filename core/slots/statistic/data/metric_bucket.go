package data

import (
	"errors"
	"math"
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

func (mb *MetricBucket) MetricEvents() []MetricEventType {
	met := make([]MetricEventType, 0, metricEventNum)
	met = append(met, MetricEventPass)
	met = append(met, MetricEventBlock)
	met = append(met, MetricEventError)
	met = append(met, MetricEventSuccess)
	met = append(met, MetricEventRt)
	return met
}

func newEmptyMetricBucket() MetricBucket {
	return MetricBucket{
		counter: make([]uint64, metricEventNum, metricEventNum),
		minRt:   math.MaxUint64,
	}
}

func (mb *MetricBucket) Add(event MetricEventType, count uint64) error {
	switch event {
	case MetricEventPass:
		mb.counter[0] += count
	case MetricEventBlock:
		mb.counter[1] += count
	case MetricEventError:
		mb.counter[2] += count
	case MetricEventSuccess:
		mb.counter[3] += count
	case MetricEventRt:
		mb.counter[4] += count
	default:
		return errors.New("unknown metric event type, " + string(event))
	}
	return nil
}

func (mb *MetricBucket) Get(event MetricEventType) (uint64, error) {
	switch event {
	case MetricEventPass:
		return mb.counter[0], nil
	case MetricEventBlock:
		return mb.counter[1], nil
	case MetricEventError:
		return mb.counter[2], nil
	case MetricEventSuccess:
		return mb.counter[3], nil
	case MetricEventRt:
		return mb.counter[4], nil
	default:
		return 0, errors.New("unknown metric event type, " + string(event))
	}
}

func (mb *MetricBucket) AddRt(rt uint64) error {
	err := mb.Add(MetricEventRt, rt)
	if rt < mb.minRt {
		mb.minRt = rt
	}
	return err
}

func (mb *MetricBucket) Reset() {
	for i := 0; i < metricEventNum; i++ {
		mb.counter[i] = 0
	}
	mb.minRt = math.MaxUint64
}
