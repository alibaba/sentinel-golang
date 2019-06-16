package data

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type WindowWrap struct {
	windowLengthInMs uint32
	WindowStart      uint64
	Value            interface{}
}

func (ww *WindowWrap) resetTo(startTime uint64) {
	ww.WindowStart = startTime
}

func (ww *WindowWrap) isTimeInWindow(timeMillis uint64) bool {
	return ww.WindowStart <= timeMillis && timeMillis < ww.WindowStart+uint64(ww.windowLengthInMs)
}

type LeapArray struct {
	windowLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	array            []*WindowWrap //实际保存的数据

	mux sync.Mutex // lock
}

func (la *LeapArray) CurrentWindow(sw BucketGenerator) (*WindowWrap, error) {
	return la.CurrentWindowWithTime(uint64(time.Now().UnixNano())/1e6, sw)
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
				WindowStart:      windowStart,
				Value:            sw.newEmptyBucket(windowStart),
			}
			la.mux.Lock()
			la.array[idx] = newWrap
			la.mux.Unlock()
			return la.array[idx], nil
		} else if windowStart == old.WindowStart {
			return old, nil
		} else if windowStart > old.WindowStart {
			// reset WindowWrap
			la.mux.Lock()
			old, _ = sw.resetWindowTo(old, windowStart)
			la.mux.Unlock()
			return old, nil
		} else if windowStart < old.WindowStart {
			// Should not go through here,
			return nil, errors.New(fmt.Sprintf("provided time timeMillis=%d is already behind old.WindowStart=%d", windowStart, old.WindowStart))
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
	return la.valuesWithTime(uint64(time.Now().UnixNano()) / 1e6)
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
				WindowStart:      uint64(time.Now().Nanosecond() / 1e6),
				Value:            newEmptyMetricBucket(),
			}
			wwp = append(wwp, wwp_)
			continue
		}
		ww := &WindowWrap{
			windowLengthInMs: wwp_.windowLengthInMs,
			WindowStart:      wwp_.WindowStart,
			Value:            wwp_.Value,
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

/**
 * The implement of sliding window based on struct MetricBucket
 */
type SlidingWindow struct {
	data       *LeapArray
	BucketType string
}

func NewSlidingWindow() *SlidingWindow {
	array_ := make([]*WindowWrap, 5)
	return &SlidingWindow{
		data: &LeapArray{
			windowLengthInMs: 200,
			sampleCount:      5,
			intervalInMs:     1000,
			array:            array_,
		},
		BucketType: "metrics",
	}
}

func (sw *SlidingWindow) newEmptyBucket(startTime uint64) interface{} {
	return newEmptyMetricBucket()
}

func (sw *SlidingWindow) resetWindowTo(ww *WindowWrap, startTime uint64) (*WindowWrap, error) {
	ww.WindowStart = startTime
	ww.Value = newEmptyMetricBucket()
	return ww, nil
}

func (sw *SlidingWindow) Count(eventType MetricEventType) uint64 {
	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		fmt.Println("sliding window fail to record success")
	}
	count := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.Value.(MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		cn := uint64(0)
		var ce error
		switch eventType {
		case MetricEventSuccess:
			cn, ce = mb.Get(MetricEventSuccess)
		case MetricEventPass:
			cn, ce = mb.Get(MetricEventPass)
		case MetricEventError:
			cn, ce = mb.Get(MetricEventError)
		case MetricEventBlock:
			cn, ce = mb.Get(MetricEventBlock)
		case MetricEventRt:
			cn, ce = mb.Get(MetricEventRt)
		default:
			ce = errors.New("unknown metric type! ")
		}
		if ce != nil {
			fmt.Println("fail to count, reason: ", ce)
		}
		count += cn
	}
	return count
}

func (sw *SlidingWindow) AddCount(eventType MetricEventType, count uint64) {
	curWindow, err := sw.data.CurrentWindow(sw)
	if err != nil || curWindow == nil || curWindow.Value == nil {
		fmt.Println("sliding window fail to record success")
		return
	}

	mb, ok := curWindow.Value.(MetricBucket)
	if !ok {
		fmt.Println("assert fail")
		return
	}

	var ae error
	switch eventType {
	case MetricEventSuccess:
		ae = mb.Add(MetricEventSuccess, count)
	case MetricEventPass:
		ae = mb.Add(MetricEventPass, count)
	case MetricEventError:
		ae = mb.Add(MetricEventError, count)
	case MetricEventBlock:
		ae = mb.Add(MetricEventBlock, count)
	case MetricEventRt:
		ae = mb.Add(MetricEventRt, count)
	default:
		ae = errors.New("unknown metric type ")
	}
	if ae != nil {
		fmt.Println("add success counter fail, reason: ", ae)
	}
}

func (sw *SlidingWindow) MaxSuccess() uint64 {

	_, err := sw.data.CurrentWindow(sw)
	if err != nil {
		fmt.Println("sliding window fail to record success")
	}

	succ := uint64(0)
	for _, ww := range sw.data.Values() {
		mb, ok := ww.Value.(MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		s, err := mb.Get(MetricEventSuccess)
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
		mb, ok := ww.Value.(MetricBucket)
		if !ok {
			fmt.Println("assert fail")
			continue
		}
		s, err := mb.Get(MetricEventSuccess)
		if err != nil {
			fmt.Println("get success counter fail, reason: ", err)
		}
		succ = uint64(math.Min(float64(succ), float64(s)))
	}
	return succ
}
