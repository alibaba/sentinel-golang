package data

import (
	"errors"
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/util"
	"math"
	"runtime"
)

type WindowWrap struct {
	windowLengthInMs uint32
	windowStart      uint64
	value            interface{}
}

func (ww *WindowWrap) resetTo(startTime uint64) {
	ww.windowStart = startTime
}

func (ww *WindowWrap) isTimeInWindow(timeMillis uint64) bool {
	return ww.windowStart <= timeMillis && timeMillis < ww.windowStart+uint64(ww.windowLengthInMs)
}

// The basic data structure of sliding windows
//
type LeapArray struct {
	windowLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	array            []*WindowWrap     //实际保存的数据
	mux              util.TriableMutex // lock
}

func (la *LeapArray) CurrentWindow(sw BucketGenerator) (*WindowWrap, error) {
	return la.CurrentWindowWithTime(util.GetTimeMilli(), sw)
}

func (la *LeapArray) CurrentWindowWithTime(timeMillis uint64, sw BucketGenerator) (*WindowWrap, error) {
	if timeMillis < 0 {
		return nil, errors.New("timeMillion is less than 0")
	}

	idx := la.calculateTimeIdx(timeMillis)
	windowStart := la.calculateStartTime(timeMillis)

	for {
		old := la.array[idx]
		if old == nil {
			newWrap := &WindowWrap{
				windowLengthInMs: la.windowLengthInMs,
				windowStart:      windowStart,
				value:            sw.newEmptyBucket(windowStart),
			}
			// must be thread safe,
			// some extreme condition,may newer override old empty WindowWrap
			// 使用cas, 确保la.array[idx]更新前是nil
			la.mux.Lock()
			if la.array[idx] == nil {
				la.array[idx] = newWrap
			}
			la.mux.Unlock()
			return la.array[idx], nil
		} else if windowStart == old.windowStart {
			return old, nil
		} else if windowStart > old.windowStart {
			// reset WindowWrap
			if la.mux.TryLock() {
				old, _ = sw.resetWindowTo(old, windowStart)
				la.mux.Unlock()
				return old, nil
			} else {
				runtime.Gosched()
			}
		} else if windowStart < old.windowStart {
			// Should not go through here,
			return nil, errors.New(fmt.Sprintf("provided time timeMillis=%d is already behind old.windowStart=%d", windowStart, old.windowStart))
		}
	}
}

func (la *LeapArray) calculateTimeIdx(timeMillis uint64) uint32 {
	timeId := (int)(timeMillis / uint64(la.windowLengthInMs))
	return uint32(timeId % len(la.array))
}

func (la *LeapArray) calculateStartTime(timeMillis uint64) uint64 {
	return timeMillis - (timeMillis % uint64(la.windowLengthInMs))
}

//  Get all the bucket in sliding window for current time;
func (la *LeapArray) Values() []*WindowWrap {
	return la.valuesWithTime(util.GetTimeMilli())
}

func (la *LeapArray) valuesWithTime(timeMillis uint64) []*WindowWrap {
	if timeMillis <= 0 {
		return nil
	}
	wwp := make([]*WindowWrap, 0)
	for _, wwp_ := range la.array {
		if wwp_ == nil {
			//fmt.Printf("current bucket is nil, index is %d \n", idx)
			wwp_ = &WindowWrap{
				windowLengthInMs: 200,
				windowStart:      util.GetTimeMilli(),
				value:            newEmptyMetricBucket(),
			}
			wwp = append(wwp, wwp_)
			continue
		}
		ww := &WindowWrap{
			windowLengthInMs: wwp_.windowLengthInMs,
			windowStart:      wwp_.windowStart,
			value:            wwp_.value,
		}
		wwp = append(wwp, ww)
	}
	return wwp
}

type BucketGenerator interface {
	// 根据开始时间，创建一个新的统计bucket, bucket的具体数据结构可以有多个
	newEmptyBucket(startTime uint64) interface{}

	// 将窗口ww重置startTime和空的统计bucket
	resetWindowTo(ww *WindowWrap, startTime uint64) (*WindowWrap, error)
}

// The implement of sliding window based on struct LeapArray
type SlidingWindow struct {
	data       *LeapArray
	BucketType string
}

func NewSlidingWindow(sampleCount uint32, intervalInMs uint32) *SlidingWindow {
	if intervalInMs%sampleCount != 0 {
		panic(fmt.Sprintf("invalid parameters, intervalInMs is %d, sampleCount is %d.", intervalInMs, sampleCount))
	}
	winLengthInMs := intervalInMs / sampleCount
	array_ := make([]*WindowWrap, 5)
	return &SlidingWindow{
		data: &LeapArray{
			windowLengthInMs: winLengthInMs,
			sampleCount:      sampleCount,
			intervalInMs:     intervalInMs,
			array:            array_,
		},
		BucketType: "metrics",
	}
}

func (sw *SlidingWindow) newEmptyBucket(startTime uint64) interface{} {
	return newEmptyMetricBucket()
}

func (sw *SlidingWindow) resetWindowTo(ww *WindowWrap, startTime uint64) (*WindowWrap, error) {
	ww.windowStart = startTime
	ww.value = newEmptyMetricBucket()
	return ww, nil
}

func (sw *SlidingWindow) Count(event MetricEventType) uint64 {
	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		fmt.Println("sliding window fail to record success")
	}
	count := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		count += mb.Get(event)
	}
	return count
}

func (sw *SlidingWindow) AddCount(event MetricEventType, count uint64) {
	curWindow, err := sw.data.CurrentWindow(sw)
	if err != nil || curWindow == nil || curWindow.value == nil {
		fmt.Println("sliding window fail to record success")
		return
	}

	mb, ok := curWindow.value.(*MetricBucket)
	if !ok {
		fmt.Println("assert fail")
		return
	}
	mb.Add(event, count)
}

func (sw *SlidingWindow) MaxSuccess() uint64 {

	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		fmt.Println("sliding window fail to record success")
	}

	succ := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		s := mb.Get(MetricEventSuccess)
		if err != nil {
			fmt.Println("get success counter fail, reason: ", err)
		}
		succ = uint64(math.Max(float64(succ), float64(s)))
	}
	return succ
}

func (sw *SlidingWindow) MinSuccess() uint64 {

	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		fmt.Println("sliding window fail to record success")
	}

	succ := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.value.(*MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		s := mb.Get(MetricEventSuccess)
		if err != nil {
			fmt.Println("get success counter fail, reason: ", err)
		}
		succ = uint64(math.Min(float64(succ), float64(s)))
	}
	return succ
}
