package base

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
)

//Test sliding windows create windows
func Test_NewBucketLeapArray(t *testing.T) {
	slidingWindow := NewBucketLeapArray(SampleCount, IntervalInMs)
	now := util.CurrentTimeMillis()

	wr, err := slidingWindow.data.currentBucketOfTime(now, slidingWindow)
	if wr == nil || wr.value.Load() == nil {
		t.Errorf("Unexcepted error")
		return
	}
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if wr.bucketStart != (now - now%uint64(WindowLengthInMs)) {
		t.Errorf("Unexcepted error, window length is not same")
	}
	if wr.value.Load() == nil {
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

	now := uint64(1976296040000) // start time is 1576296044500, [1576296040000, 1576296050000]

	start := util.CurrentTimeMillis()
	var cnt = uint64(0)
	for i := 0; i < GoroutineNum; i++ {
		go coroutineTask(wg, slidingWindow, now, &cnt)
	}
	wg.Wait()
	t.Logf("Finish goroutines:  %d", atomic.LoadUint64(&cnt))
	end := util.CurrentTimeMillis()
	t.Logf("Finish %d goroutines, cost time is %d ns \n", atomic.LoadUint64(&cnt), end-start)

	success := slidingWindow.CountWithTime(now+9999, base.MetricEventComplete)
	pass := slidingWindow.CountWithTime(now+9999, base.MetricEventPass)
	block := slidingWindow.CountWithTime(now+9999, base.MetricEventBlock)
	errNum := slidingWindow.CountWithTime(now+9999, base.MetricEventError)
	rt := slidingWindow.CountWithTime(now+9999, base.MetricEventRt)
	if success == GoroutineNum && pass == GoroutineNum && block == GoroutineNum && errNum == GoroutineNum && rt == GoroutineNum*10 {
		t.Logf("Success %d, pass %d, block %d, error %d, rt %d\n", success, pass, block, errNum, rt)
	} else {
		t.Logf("Success %d, pass %d, block %d, error %d, rt %d\n", success, pass, block, errNum, rt)
		t.Errorf("Concurrency error\n")
	}
}

func coroutineTask(wg *sync.WaitGroup, slidingWindow *BucketLeapArray, now uint64, counter *uint64) {
	inc := rand.Uint64() % 10000
	slidingWindow.addCountWithTime(now+inc, base.MetricEventComplete, 1)
	slidingWindow.addCountWithTime(now+inc, base.MetricEventPass, 1)
	slidingWindow.addCountWithTime(now+inc, base.MetricEventBlock, 1)
	slidingWindow.addCountWithTime(now+inc, base.MetricEventError, 1)
	slidingWindow.addCountWithTime(now+inc, base.MetricEventRt, 10)

	atomic.AddUint64(counter, 1)
	wg.Done()
}

func TestBucketLeapArray_resetWindowTo(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	idx := 6
	oldWindow := bla.data.array.get(idx)
	oldBucket := oldWindow.value.Load()
	if oldBucket == nil {
		t.Errorf("BucketLeapArray init error.")
	}
	bucket, ok := oldBucket.(*MetricBucket)
	if !ok {
		t.Errorf("Fail to assert bucket to MetricBucket.")
	}
	bucket.Add(base.MetricEventPass, 100)
	bucket.Add(base.MetricEventBlock, 100)

	wantStartTime := util.CurrentTimeMillis() + 1000
	got := bla.resetWindowTo(oldWindow, wantStartTime)
	newBucket := got.value.Load()
	if newBucket == nil {
		t.Errorf("got window is nil.")
	}
	newRealBucket, ok := newBucket.(*MetricBucket)
	if !ok {
		t.Errorf("Fail to assert bucket to MetricBucket.")
	}
	if newRealBucket.Get(base.MetricEventPass) != 0 {
		t.Errorf("BucketLeapArray.resetWindowTo() execute fail.")
	}
	if newRealBucket.Get(base.MetricEventBlock) != 0 {
		t.Errorf("BucketLeapArray.resetWindowTo() execute fail.")
	}
}
