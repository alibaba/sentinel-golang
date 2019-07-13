package flow

import (
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"github.com/sentinel-group/sentinel-golang/core/util"
	"math"
	"sync/atomic"
	"time"
)

type TrafficShapingController interface {
	CanPass(ctx *base.Context, node *base.DefaultNode, acquire uint32) bool
}

type DefaultController struct {
	grade FlowGradeType
	count uint64
}

func (dc *DefaultController) CanPass(ctx *base.Context, node *base.DefaultNode, acquire uint32) bool {
	curCount := dc.avgUsedTokens(node)
	if (curCount + uint64(acquire)) > dc.count {
		return false
	}
	return true
}

func (dc *DefaultController) avgUsedTokens(node *base.DefaultNode) uint64 {
	if node == nil {
		return 0
	}
	if dc.grade == FlowGradeThread {
		return node.CurGoroutineNum()
	}
	return node.PassQps()
}

type RateLimiterController struct {
	count             uint64
	maxQueueingTimeMs int64
	latestPassedTime  int64
}

func (rlc *RateLimiterController) CanPass(ctx *base.Context, node *base.DefaultNode, acquire uint32) bool {
	if acquire < 0 {
		return true
	}
	// Reject when count is less or equal than 0.
	// Otherwise,the costTime will be max of long and waitTime will overflow in some cases.
	currentTime := int64(util.GetTimeMilli())
	// Calculate the interval between every two requests.
	costTime := int64(math.Round(float64(uint64(acquire)/rlc.count) * 1000))
	// Expected pass time of this request.
	expectedTime := costTime + atomic.LoadInt64(&rlc.latestPassedTime)

	if expectedTime < currentTime {
		// Contention may exist here, but it's okay.
		atomic.CompareAndSwapInt64(&rlc.latestPassedTime, rlc.latestPassedTime, currentTime)
		return true
	} else {
		// Calculate the time to wait.
		waitTime := costTime + atomic.LoadInt64(&rlc.latestPassedTime) - int64(util.GetTimeMilli())
		if waitTime > rlc.maxQueueingTimeMs {
			atomic.AddInt64(&rlc.latestPassedTime, -costTime)
			return false
		}
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
		return true
	}
}

type WarmUpController struct {
}

func (wpc WarmUpController) CanPass(ctx *base.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}

type WarmUpRateLimiterController struct {
}

func (wpc WarmUpRateLimiterController) CanPass(ctx *base.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}
