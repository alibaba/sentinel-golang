package flow

import (
	"math"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/util"
)

// ThrottlingChecker limits the time interval between two requests.
type ThrottlingChecker struct {
	owner             *TrafficShapingController
	maxQueueingTimeNs uint64
	statIntervalNs    uint64
	lastPassedTime    uint64
}

func NewThrottlingChecker(owner *TrafficShapingController, timeoutMs uint32, statIntervalMs uint32) *ThrottlingChecker {
	var statIntervalNs uint64
	if statIntervalMs == 0 {
		defaultIntervalMs := config.MetricStatisticIntervalMs()
		if defaultIntervalMs == 0 {
			defaultIntervalMs = 1000
		}
		statIntervalNs = uint64(defaultIntervalMs) * util.UnixTimeUnitOffset
	} else {
		statIntervalNs = uint64(statIntervalMs) * util.UnixTimeUnitOffset
	}
	return &ThrottlingChecker{
		owner:             owner,
		maxQueueingTimeNs: uint64(timeoutMs) * util.UnixTimeUnitOffset,
		statIntervalNs:    statIntervalNs,
		lastPassedTime:    0,
	}
}
func (c *ThrottlingChecker) BoundOwner() *TrafficShapingController {
	return c.owner
}

func (c *ThrottlingChecker) DoCheck(_ base.StatNode, batchCount uint32, threshold float64) *base.TokenResult {
	// Pass when batch count is less or equal than 0.
	if batchCount <= 0 {
		return nil
	}

	var rule *Rule
	if c.BoundOwner() != nil {
		rule = c.BoundOwner().BoundRule()
	}

	if threshold <= 0.0 {
		msg := "flow throttling check not pass, threshold is <= 0.0"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, rule, nil)
	}
	if float64(batchCount) > threshold {
		return base.NewTokenResultBlocked(base.BlockTypeFlow)
	}
	// Here we use nanosecond so that we could control the queueing time more accurately.
	curNano := util.CurrentTimeNano()

	// The interval between two requests (in nanoseconds).
	intervalNs := uint64(math.Ceil(float64(batchCount) / threshold * float64(c.statIntervalNs)))

	// Expected pass time of this request.
	expectedTime := atomic.LoadUint64(&c.lastPassedTime) + intervalNs
	if expectedTime <= curNano {
		// Contention may exist here, but it's okay.
		atomic.StoreUint64(&c.lastPassedTime, curNano)
		return nil
	}

	estimatedQueueingDuration := atomic.LoadUint64(&c.lastPassedTime) + intervalNs - util.CurrentTimeNano()
	if estimatedQueueingDuration > c.maxQueueingTimeNs {
		msg := "flow throttling check not pass, estimated queueing time exceeds max queueing time"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, rule, nil)
	}

	oldTime := atomic.AddUint64(&c.lastPassedTime, intervalNs)
	estimatedQueueingDuration = oldTime - util.CurrentTimeNano()
	if estimatedQueueingDuration > c.maxQueueingTimeNs {
		// Subtract the interval.
		atomic.AddUint64(&c.lastPassedTime, ^(intervalNs - 1))
		msg := "flow throttling check not pass, estimated queueing time exceeds max queueing time"
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, msg, rule, nil)
	}
	if estimatedQueueingDuration > 0 {
		return base.NewTokenResultShouldWait(estimatedQueueingDuration / util.UnixTimeUnitOffset)
	} else {
		return base.NewTokenResultShouldWait(0)
	}
}
