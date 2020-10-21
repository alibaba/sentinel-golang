package flow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
)

func newTestThrottlingChecker(intervalMs uint32, threshold float64, timeoutMs uint32) (*ThrottlingChecker, error) {
	rule := &Rule{
		Resource:               "test",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Throttling,
		Threshold:              threshold,
		StatIntervalInMs:       intervalMs,
		MaxQueueingTimeMs:      timeoutMs,
	}

	boundStat, err := generateStatFor(rule)
	if err != nil {
		return nil, err
	}

	tsc, err := NewTrafficShapingController(rule, boundStat)
	if err != nil || tsc == nil {
		return nil, err
	}

	return NewThrottlingChecker(tsc, timeoutMs), nil
}

func TestThrottlingChecker_DoCheckNoQueueingSingleThread(t *testing.T) {
	intervalMs := 10000
	threshold := 50.0
	timeoutMs := 0

	tc, err := newTestThrottlingChecker(uint32(intervalMs), threshold, uint32(timeoutMs))
	assert.True(t, err == nil, err)

	// The first request will pass.
	ret := tc.DoCheck(nil, 1, threshold)
	assert.True(t, ret == nil || ret.IsPass())

	reqCount := 10
	for i := 0; i < reqCount; i++ {
		assert.True(t, tc.DoCheck(nil, 1, threshold).IsBlocked())
	}
	time.Sleep(time.Duration(intervalMs/int(threshold)*reqCount+10) * time.Millisecond)

	assert.True(t, tc.DoCheck(nil, 1, threshold) == nil)
	assert.True(t, tc.DoCheck(nil, 1, threshold).IsBlocked())
}

func TestThrottlingChecker_DoCheckSingleThread(t *testing.T) {
	intervalMs := 10000
	threshold := 50.0
	timeoutMs := 2000

	tc, err := newTestThrottlingChecker(uint32(intervalMs), threshold, uint32(timeoutMs))
	assert.True(t, err == nil, err)

	resultList := make([]*base.TokenResult, 0)
	reqCount := 20
	for i := 0; i < reqCount; i++ {
		res := tc.DoCheck(nil, 1, threshold)
		resultList = append(resultList, res)
	}
	assert.True(t, resultList[0] == nil)

	// waitCount is count of request that will wait and not be blocked
	waitCount := int(float64(timeoutMs) / (float64(intervalMs) / threshold))
	for i := 1; i <= waitCount; i++ {
		assert.True(t, resultList[i].Status() == base.ResultStatusShouldWait)
		wt := resultList[i].WaitMs()
		assert.InEpsilon(t, i*1000/int(waitCount), wt, 10)
	}
	for i := waitCount + 1; i < reqCount; i++ {
		assert.True(t, resultList[i].IsBlocked())
	}
}

func TestThrottlingChecker_DoCheckQueueingParallel(t *testing.T) {
	intervalMs := 10000
	threshold := 50.0
	timeoutMs := 0

	tc, err := newTestThrottlingChecker(uint32(intervalMs), threshold, uint32(timeoutMs))
	assert.True(t, err == nil, err)

	assert.True(t, tc.DoCheck(nil, 1, threshold) == nil)

	wg := &sync.WaitGroup{}
	gc := 24
	wg.Add(gc)

	var waitCount, blockCount int32 = 0, 0
	for i := 0; i < gc; i++ {
		go func() {
			res := tc.DoCheck(nil, 1, threshold)
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
	assert.InEpsilon(t, threshold/(float64(intervalMs)/1000.0), waitCount, 1)
}
