package flow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
)

func TestThrottlingChecker_DoCheckNoQueueingSingleThread(t *testing.T) {
	tc := NewThrottlingChecker(0)
	var qps float64 = 5

	// The first request will pass.
	ret := tc.DoCheck(nil, 1, qps)
	assert.True(t, ret == nil || ret.IsPass())

	for i := 0; i < int(qps); i++ {
		assert.True(t, tc.DoCheck(nil, 1, qps).IsBlocked())
	}
	time.Sleep(time.Duration(1000/int(qps)+10) * time.Millisecond)

	assert.True(t, tc.DoCheck(nil, 1, qps) == nil)
	assert.True(t, tc.DoCheck(nil, 1, qps).IsBlocked())
}

func TestThrottlingChecker_DoCheckSingleThread(t *testing.T) {
	tc := NewThrottlingChecker(1000)
	var qps float64 = 5
	resultList := make([]*base.TokenResult, 0)
	for i := 0; i < 10; i++ {
		res := tc.DoCheck(nil, 1, qps)
		resultList = append(resultList, res)
	}
	assert.True(t, resultList[0] == nil)

	for i := 1; i <= int(qps); i++ {
		assert.True(t, resultList[i].Status() == base.ResultStatusShouldWait)
		wt := resultList[i].WaitMs()
		assert.InEpsilon(t, i*1000/int(qps), wt, 10)
	}
	for i := int(qps) + 1; i < 10; i++ {
		assert.True(t, resultList[i].IsBlocked())
	}
}

func TestThrottlingChecker_DoCheckQueueingParallel(t *testing.T) {
	tc := NewThrottlingChecker(1000)
	var qps float64 = 5

	assert.True(t, tc.DoCheck(nil, 1, qps) == nil)

	wg := &sync.WaitGroup{}
	gc := 24
	wg.Add(gc)

	var waitCount, blockCount int32 = 0, 0
	for i := 0; i < gc; i++ {
		go func() {
			res := tc.DoCheck(nil, 1, qps)
			if res.IsBlocked() {
				atomic.AddInt32(&blockCount, 1)
			}
			if res.Status() == base.ResultStatusShouldWait {
				atomic.AddInt32(&waitCount, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(gc), waitCount+blockCount)
	// Non-strict mode may not be strictly accurate, so here we tolerate a delta.
	assert.InEpsilon(t, qps, waitCount, 1)
}
