package base

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

//Test sliding windows create windows
func Test_NewBucketLeapArray(t *testing.T) {
	slidingWindow := NewBucketLeapArray(SampleCount, IntervalInMs)
	now := util.CurrentTimeMillis()

	wr, err := slidingWindow.data.currentWindowWithTime(now, slidingWindow)
	if wr == nil || wr.value == nil {
		t.Errorf("Unexcepted error")
	}
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr.windowStart != (now - now%uint64(WindowLengthInMs)) {
		t.Errorf("Unexcepted error, window length is not same")
	}
	if wr.value == nil {
		t.Errorf("Unexcepted error, value is nil")
	}
	if slidingWindow.Count(base.MetricEventPass) != 0 {
		t.Errorf("Unexcepted error, pass value is invalid")
	}
}

func Test_UpdateBucket_Concurrent(t *testing.T) {
	slidingWindow := NewBucketLeapArray(SampleCount, IntervalInMs)

	const GoroutineNum = 3000
	wg := &sync.WaitGroup{}
	wg.Add(GoroutineNum)
	start := util.CurrentTimeMillis()
	var cnt = uint64(0)
	for i := 0; i < GoroutineNum; i++ {
		go coroutineTask(wg, slidingWindow, &cnt)
	}
	wg.Wait()
	t.Logf("Finish goroutines:  %d", atomic.LoadUint64(&cnt))
	end := util.CurrentTimeMillis()
	t.Logf("Finish %d goroutines, cost time is %d ns \n", atomic.LoadUint64(&cnt), (end - start))
	success := slidingWindow.Count(base.MetricEventComplete)
	pass := slidingWindow.Count(base.MetricEventPass)
	block := slidingWindow.Count(base.MetricEventBlock)
	errNum := slidingWindow.Count(base.MetricEventError)
	rt := slidingWindow.Count(base.MetricEventRt)
	if success == GoroutineNum && pass == GoroutineNum && block == GoroutineNum && errNum == GoroutineNum && rt == GoroutineNum*10 {
		t.Logf("Success %d, pass %d, block %d, error %d, rt %d\n", success, pass, block, errNum, rt)
	} else {
		t.Errorf("Concurrency error\n")
	}
}

func coroutineTask(wg *sync.WaitGroup, slidingWindow *BucketLeapArray, counter *uint64) {
	time.Sleep(time.Millisecond * 3)
	slidingWindow.AddCount(base.MetricEventComplete, 1)
	time.Sleep(time.Millisecond * 3)
	slidingWindow.AddCount(base.MetricEventPass, 1)
	time.Sleep(time.Millisecond * 3)
	slidingWindow.AddCount(base.MetricEventBlock, 1)
	time.Sleep(time.Millisecond * 3)
	slidingWindow.AddCount(base.MetricEventError, 1)
	time.Sleep(time.Millisecond * 3)
	slidingWindow.AddCount(base.MetricEventRt, 10)
	atomic.AddUint64(counter, 1)
	time.Sleep(time.Millisecond * 3)
	wg.Done()
}
