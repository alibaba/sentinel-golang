package data

import (
	"sync"
	"sync/atomic"
	"testing"
	time2 "time"
)

const (
	windowLengthImMs_ uint32 = 200
	sampleCount_      uint32 = 5
	intervalInMs_     uint32 = 1000
)

//Test sliding windows create windows
func TestNewWindow(t *testing.T) {
	slidingWindow := NewSlidingWindow(sampleCount_, intervalInMs_)
	time := uint64(time2.Now().UnixNano() / 1e6)

	wr, err := slidingWindow.data.CurrentWindowWithTime(time, slidingWindow)
	if wr == nil {
		t.Errorf("Unexcepted error")
	}
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr.windowLengthInMs != windowLengthImMs_ {
		t.Errorf("Unexcepted error, winlength is not same")
	}
	if wr.windowStart != (time - time%uint64(windowLengthImMs_)) {
		t.Errorf("Unexcepted error, winlength is not same")
	}
	if wr.value == nil {
		t.Errorf("Unexcepted error, value is nil")
	}
	if slidingWindow.Count(MetricEventPass) != 0 {
		t.Errorf("Unexcepted error, pass value is invalid")
	}
}

// Test the logic get window start time.
func TestLeapArrayWindowStart(t *testing.T) {
	slidingWindow := NewSlidingWindow(sampleCount_, intervalInMs_)
	firstTime := uint64(time2.Now().UnixNano() / 1e6)
	previousWindowStart := firstTime - firstTime%uint64(windowLengthImMs_)

	wr, err := slidingWindow.data.CurrentWindowWithTime(firstTime, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr.windowLengthInMs != windowLengthImMs_ {
		t.Errorf("Unexpected error, winLength is not same")
	}
	if wr.windowStart != previousWindowStart {
		t.Errorf("Unexpected error, winStart is not same")
	}
}

// test sliding window has multi windows
func TestWindowAfterOneInterval(t *testing.T) {
	slidingWindow := NewSlidingWindow(sampleCount_, intervalInMs_)
	firstTime := uint64(time2.Now().UnixNano() / 1e6)
	previousWindowStart := firstTime - firstTime%uint64(windowLengthImMs_)

	wr, err := slidingWindow.data.CurrentWindowWithTime(firstTime, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr.windowLengthInMs != windowLengthImMs_ {
		t.Errorf("Unexpected error, winLength is not same")
	}
	if wr.windowStart != previousWindowStart {
		t.Errorf("Unexpected error, winStart is not same")
	}
	if wr.value == nil {
		t.Errorf("Unexcepted error")
	}
	mb, ok := wr.value.(MetricBucket)
	if !ok {
		t.Errorf("Unexcepted error")
	}
	mb.Add(MetricEventPass, 1)
	mb.Add(MetricEventBlock, 1)
	mb.Add(MetricEventSuccess, 1)
	mb.Add(MetricEventError, 1)

	if mb.Get(MetricEventPass) != 1 {
		t.Errorf("Unexcepted error")
	}
	if mb.Get(MetricEventBlock) != 1 {
		t.Errorf("Unexcepted error")
	}
	if mb.Get(MetricEventSuccess) != 1 {
		t.Errorf("Unexcepted error")
	}
	if mb.Get(MetricEventError) != 1 {
		t.Errorf("Unexcepted error")
	}

	middleTime := previousWindowStart + uint64(windowLengthImMs_)/2
	wr2, err := slidingWindow.data.CurrentWindowWithTime(middleTime, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr2.windowStart != previousWindowStart {
		t.Errorf("Unexpected error, winStart is not same")
	}
	mb2, ok := wr2.value.(MetricBucket)
	if !ok {
		t.Errorf("Unexcepted error")
	}
	if wr != wr2 {
		t.Errorf("Unexcepted error")
	}
	mb2.Add(MetricEventPass, 1)
	if mb.Get(MetricEventPass) != 2 {
		t.Errorf("Unexcepted error")
	}
	if mb.Get(MetricEventBlock) != 1 {
		t.Errorf("Unexcepted error")
	}

	lastTime := middleTime + uint64(windowLengthImMs_)/2
	wr3, err := slidingWindow.data.CurrentWindowWithTime(lastTime, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr3.windowLengthInMs != windowLengthImMs_ {
		t.Errorf("Unexpected error")
	}
	if (wr3.windowStart - uint64(windowLengthImMs_)) != previousWindowStart {
		t.Errorf("Unexpected error")
	}
	mb3, ok := wr3.value.(MetricBucket)
	if !ok {
		t.Errorf("Unexcepted error")
	}
	if &mb3 == nil {
		t.Errorf("Unexcepted error")
	}

	if mb3.Get(MetricEventPass) != 0 {
		t.Errorf("Unexcepted error")
	}
	if mb3.Get(MetricEventBlock) != 0 {
		t.Errorf("Unexcepted error")
	}
}

func TestNTimeMultiGoroutineUpdateEmptyWindow(t *testing.T) {
	for i := 0; i < 1000; i++ {
		_nTestMultiGoroutineUpdateEmptyWindow(t)
	}
}

func _task(wg *sync.WaitGroup, slidingWindow *SlidingWindow, ti uint64, t *testing.T, ct *uint64) {
	wr, err := slidingWindow.data.CurrentWindowWithTime(ti, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	mb, ok := wr.value.(MetricBucket)
	if !ok {
		t.Errorf("Unexcepted error")
	}
	mb.Add(MetricEventPass, 1)
	mb.Add(MetricEventBlock, 1)
	mb.Add(MetricEventSuccess, 1)
	mb.Add(MetricEventError, 1)
	atomic.AddUint64(ct, 1)
	wg.Done()
}

func _nTestMultiGoroutineUpdateEmptyWindow(t *testing.T) {
	slidingWindow := NewSlidingWindow(sampleCount_, intervalInMs_)
	firstTime := uint64(time2.Now().UnixNano() / 1e6)

	const GoroutineNum = 10000
	wg := &sync.WaitGroup{}
	wg.Add(GoroutineNum)
	st := time2.Now().UnixNano()
	var cnt = uint64(0)
	for i := 0; i < GoroutineNum; i++ {
		go _task(wg, slidingWindow, firstTime, t, &cnt)
	}
	wg.Wait()
	t.Logf("finish goroutines:  %d", atomic.LoadUint64(&cnt))
	et := time2.Now().UnixNano()
	dif := et - st
	t.Logf("finish all goroutines, cost time is %d", dif)
	wr2, err := slidingWindow.data.CurrentWindowWithTime(firstTime, slidingWindow)
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	mb2, ok := wr2.value.(MetricBucket)
	if !ok {
		t.Errorf("Unexcepted error")
	}
	if mb2.Get(MetricEventPass) != GoroutineNum {
		t.Errorf("Unexcepted error, infact, %d", mb2.Get(MetricEventPass))
	}
	if mb2.Get(MetricEventBlock) != GoroutineNum {
		t.Errorf("Unexcepted error, infact, %d", mb2.Get(MetricEventBlock))
	}
}
