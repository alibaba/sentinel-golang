package base

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/logging"
	"sync/atomic"
	"unsafe"
)

var logger = logging.GetDefaultLogger()

// The implement of sliding window based on leapArray and metricBucket
// metricBucket is used to record statistic metrics
// Default, BucketLeapArray is per resource
type BucketLeapArray struct {
	data          leapArray
	LeapArrayType string
}

// sampleCount is the number of slots
// intervalInMs is the time length of sliding window
func NewBucketLeapArray(sampleCount uint32, intervalInMs uint32) *BucketLeapArray {
	if intervalInMs%sampleCount != 0 {
		panic(fmt.Sprintf("Invalid parameters, intervalInMs is %d, sampleCount is %d.", intervalInMs, sampleCount))
	}
	winLengthInMs := intervalInMs / sampleCount
	ret := &BucketLeapArray{
		data: leapArray{
			windowLengthInMs: winLengthInMs,
			sampleCount:      sampleCount,
			intervalInMs:     intervalInMs,
			array:            nil,
		},
		LeapArrayType: "metricBucket",
	}
	arr := newAtomicWindowWrapArray(int(sampleCount), winLengthInMs, ret)
	ret.data.array = arr
	return ret
}

func (bla *BucketLeapArray) SampleCount() uint32 {
	return bla.data.sampleCount
}

func (bla *BucketLeapArray) IntervalInMs() uint32 {
	return bla.data.intervalInMs
}

func (bla *BucketLeapArray) WindowLengthInMs() uint32 {
	return bla.data.windowLengthInMs
}

func (bla *BucketLeapArray) GetIntervalInSecond() float64 {
	return float64(bla.IntervalInMs()) / 1000.0
}

func (bla *BucketLeapArray) newEmptyBucket() interface{} {
	return newMetricBucket()
}

func (bla *BucketLeapArray) resetWindowTo(ww *windowWrap, startTime uint64) *windowWrap {
	atomic.StoreUint64(&ww.windowStart, startTime)
	oldValP, ok := ww.value.(*metricBucket)
	if !ok {
		panic("windowWrap value assert fail.")
	}
	oldValPtr := unsafe.Pointer(oldValP)
	atomic.StorePointer(&oldValPtr, unsafe.Pointer(newMetricBucket()))
	return ww
}

// Write method
// It might panic
func (bla *BucketLeapArray) AddCount(event base.MetricEvent, count int64) {
	curWindow, err := bla.data.currentWindow(bla)
	if err != nil {
		logger.Errorf("Fail to get current window, err: %+v.", errors.WithStack(err))
		return
	}
	if curWindow == nil || curWindow.value == nil {
		logger.Error("Current window is nil.")
		return
	}
	mb, ok := curWindow.value.(*metricBucket)
	if !ok {
		logger.Error("Fail to assert metricBucket type.")
		return
	}
	mb.add(event, count)
}

// For test
func (bla *BucketLeapArray) AddCountWithTime(now uint64, event base.MetricEvent, count int64) {
	curWindow, err := bla.data.currentWindowWithTime(now, bla)
	if err != nil {
		logger.Errorf("Fail to get current window, err: %+v.", errors.WithStack(err))
		return
	}
	if curWindow == nil || curWindow.value == nil {
		logger.Error("Current window is nil.")
		return
	}
	mb, ok := curWindow.value.(*metricBucket)
	if !ok {
		logger.Error("Fail to assert metricBucket type.")
		return
	}
	mb.add(event, count)
}


// Read method, need to adapt upper application
// it might panic
func (bla *BucketLeapArray) Count(event base.MetricEvent) int64 {
	_, err := bla.data.currentWindow(bla)
	if err != nil {
		logger.Errorf("Fail to get current window, err: %+v.", errors.WithStack(err))
	}
	count := int64(0)
	for _, ww := range bla.data.values() {
		mb, ok := ww.value.(*metricBucket)
		if !ok {
			logger.Error("Fail to assert metricBucket type.")
			continue
		}
		count += mb.get(event)
	}
	return count
}

func (bla *BucketLeapArray) CountWithTime(now uint64, event base.MetricEvent) int64 {
	_, err := bla.data.currentWindowWithTime(now, bla)
	if err != nil {
		logger.Errorf("Fail to get current window, err: %+v.", errors.WithStack(err))
	}
	count := int64(0)
	for _, ww := range bla.data.values() {
		mb, ok := ww.value.(*metricBucket)
		if !ok {
			logger.Error("Fail to assert metricBucket type.")
			continue
		}
		count += mb.get(event)
	}
	return count
}


// Read method, get all windowWrap.
func (bla *BucketLeapArray) Values(now uint64) []*windowWrap {
	_, err := bla.data.currentWindowWithTime(now, bla)
	if err != nil {
		logger.Errorf("Fail to get current(%d) window, err: %+v.", now, errors.WithStack(err))
	}
	return bla.data.valuesWithTime(now)
}

func (bla *BucketLeapArray) ValuesWithConditional(now uint64, predicate base.TimePredicate) []*windowWrap {
	_, err := bla.data.currentWindowWithTime(now, bla)
	if err != nil {
		logger.Errorf("Fail to get current(%d) window, err: %+v.", now, errors.WithStack(err))
	}
	return bla.data.ValuesWithConditional(now, predicate)
}


func (bla *BucketLeapArray) MinRt() int64 {
	_, err := bla.data.currentWindow(bla)
	if err != nil {
		logger.Errorf("Fail to get current window, err: %+v.", errors.WithStack(err))
	}

	ret := base.DefaultStatisticMaxRt

	for _, v := range bla.data.values() {
		w, ok := v.value.(*metricBucket)
		if !ok {
			logger.Errorf("Fail to assert windowWrap.value(%+v) to metricBucket.", v.value)
			continue
		}
		mr := w.getMinRt()
		if ret > mr {
			ret = mr
		}
	}
	return ret
}
