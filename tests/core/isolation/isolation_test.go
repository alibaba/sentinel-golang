package isolation

import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrency(t *testing.T) {
	var (
		resource      = "test"
		maxConcurrent = 20000
		sleepChan     = make(chan struct{})
		passNum       = int32(0) // number of goroutine successfully passed
		isContinue    = true
	)
	_, _ = isolation.LoadRules([]*isolation.Rule{
		{
			Resource:   resource,
			MetricType: isolation.Concurrency,
			Threshold:  uint32(maxConcurrent),
		},
	})
	_, _ = circuitbreaker.LoadRules([]*circuitbreaker.Rule{
		{
			Resource:         resource,
			Strategy:         circuitbreaker.ErrorRatio,
			RetryTimeoutMs:   uint32(5000),
			StatIntervalMs:   5000,
			Threshold:        0.5,
		},
	})
	for isContinue {
		go func() {
			entry, err := sentinel.Entry(resource)
			if err != nil {
				isContinue = false
			}else {
				defer entry.Exit()
				// when passing, passNum++
				atomic.AddInt32(&passNum, 1)
				// suppose to block here
				<-sleepChan
			}
		}()
	}
	time.Sleep(1 * time.Second)
	if passNum != int32(maxConcurrent) {
		t.Fatalf("current concurrent is %d, set max concurrent is %d, this is not equal", passNum, maxConcurrent)
	}
	t.Logf("current concurrent is %d, set max concurrent is %d, this is equal", passNum, maxConcurrent)
}