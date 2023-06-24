// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

// Test sliding windows create buckets
func Test_NewBucketLeapArray(t *testing.T) {
	slidingWindow := NewBucketLeapArray(SampleCount, IntervalInMs)
	now := util.CurrentTimeMillis()

	br, err := slidingWindow.data.currentBucketOfTime(now, slidingWindow)
	if br == nil || br.Value.Load() == nil {
		t.Errorf("Unexcepted error")
		return
	}
	if err != nil {
		t.Errorf("Unexcepted error")
	}
	if br.BucketStart != (now - now%uint64(BucketLengthInMs)) {
		t.Errorf("Unexcepted error, bucket length is not same")
	}
	if br.Value.Load() == nil {
		t.Errorf("Unexcepted error, Value is nil")
	}
	if slidingWindow.Count(base.MetricEventPass) != 0 {
		t.Errorf("Unexcepted error, pass Value is invalid")
	}
}

func Test_UpdateBucket_Concurrent(t *testing.T) {
	slidingWindow := NewBucketLeapArray(SampleCount, IntervalInMs)

	const GoroutineNum = 3000
	now := uint64(1976296040000) // start time is 1576296044500, [1576296040000, 1576296050000]

	var cnt = uint64(0)
	for t := now; t < now+uint64(IntervalInMs); {
		slidingWindow.addCountWithTime(t, base.MetricEventComplete, 1)
		slidingWindow.addCountWithTime(t, base.MetricEventPass, 1)
		slidingWindow.addCountWithTime(t, base.MetricEventBlock, 1)
		slidingWindow.addCountWithTime(t, base.MetricEventError, 1)
		slidingWindow.addCountWithTime(t, base.MetricEventRt, 10)
		t = t + 500
	}
	for _, b := range slidingWindow.Values(uint64(1976296049500)) {
		bucket, ok := b.Value.Load().(*MetricBucket)
		assert.True(t, ok)
		assert.True(t, bucket.Get(base.MetricEventComplete) == 1)
		assert.True(t, bucket.Get(base.MetricEventPass) == 1)
		assert.True(t, bucket.Get(base.MetricEventBlock) == 1)
		assert.True(t, bucket.Get(base.MetricEventError) == 1)
		assert.True(t, bucket.Get(base.MetricEventRt) == 10)
	}

	wg := &sync.WaitGroup{}
	wg.Add(GoroutineNum - 20)
	for i := 0; i < GoroutineNum-20; i++ {
		go coroutineTask(wg, slidingWindow, now, &cnt)
	}
	wg.Wait()

	success := slidingWindow.CountWithTime(now+9999, base.MetricEventComplete)
	pass := slidingWindow.CountWithTime(now+9999, base.MetricEventPass)
	block := slidingWindow.CountWithTime(now+9999, base.MetricEventBlock)
	errNum := slidingWindow.CountWithTime(now+9999, base.MetricEventError)
	rt := slidingWindow.CountWithTime(now+9999, base.MetricEventRt)
	if success == GoroutineNum && pass == GoroutineNum && block == GoroutineNum && errNum == GoroutineNum && rt == GoroutineNum*10 {
		fmt.Printf("Success %d, pass %d, block %d, error %d, rt %d\n", success, pass, block, errNum, rt)
	} else {
		fmt.Printf("Success %d, pass %d, block %d, error %d, rt %d\n", success, pass, block, errNum, rt)
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

func TestBucketLeapArray_resetBucketTo(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	idx := 19
	oldBucketWrap := bla.data.array.get(idx)
	oldBucket := oldBucketWrap.Value.Load()
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
	got := bla.ResetBucketTo(oldBucketWrap, wantStartTime)
	newBucket := got.Value.Load()
	if newBucket == nil {
		t.Errorf("got bucket is nil.")
	}
	newRealBucket, ok := newBucket.(*MetricBucket)
	if !ok {
		t.Errorf("Fail to assert bucket to MetricBucket.")
	}
	if newRealBucket.Get(base.MetricEventPass) != 0 {
		t.Errorf("BucketLeapArray.ResetBucketTo() execute fail.")
	}
	if newRealBucket.Get(base.MetricEventBlock) != 0 {
		t.Errorf("BucketLeapArray.ResetBucketTo() execute fail.")
	}
}

func TestAddCount(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	bla.AddCount(base.MetricEventPass, 1)
	passCount := bla.Count(base.MetricEventPass)
	assert.True(t, passCount == 1)
}

func TestUpdateConcurrency(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	bla.UpdateConcurrency(1)
	bla.UpdateConcurrency(3)
	bla.UpdateConcurrency(2)
	mc := bla.MaxConcurrency()
	assert.True(t, mc == 3)
}

func TestMinRt(t *testing.T) {
	t.Run("TestMinRt_Default", func(t *testing.T) {
		bla := NewBucketLeapArray(SampleCount, IntervalInMs)
		minRt := bla.MinRt()
		assert.True(t, minRt == base.DefaultStatisticMaxRt)
	})

	t.Run("TestMinRt", func(t *testing.T) {
		bla := NewBucketLeapArray(SampleCount, IntervalInMs)
		bla.AddCount(base.MetricEventRt, 100)
		minRt := bla.MinRt()
		assert.True(t, minRt == 100)
	})
}

func TestGetIntervalInSecond(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	second := bla.GetIntervalInSecond()
	assert.True(t, util.Float64Equals(float64(IntervalInMs)/1000.0, second))
}

func TestDataType(t *testing.T) {
	bla := NewBucketLeapArray(SampleCount, IntervalInMs)
	dataType := bla.DataType()
	assert.True(t, dataType == "MetricBucket")
}
