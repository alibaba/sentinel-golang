package hotspot

import (
	"fmt"
	"math"
	"runtime"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/hotspot/cache"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

type TrafficShapingController interface {
	PerformChecking(arg interface{}, acquireCount int64) *base.TokenResult

	BoundParamIndex() int

	BoundMetric() *ParamsMetric

	BoundRule() *Rule
}

type baseTrafficShapingController struct {
	r *Rule

	res           string
	metricType    MetricType
	paramIndex    int
	threshold     float64
	specificItems map[interface{}]int64
	durationInSec int64

	metric *ParamsMetric
}

func newBaseTrafficShapingControllerWithMetric(r *Rule, metric *ParamsMetric) *baseTrafficShapingController {
	specificItems := parseSpecificItems(r.SpecificItems)
	return &baseTrafficShapingController{
		r:             r,
		res:           r.Resource,
		metricType:    r.MetricType,
		paramIndex:    r.ParamIndex,
		threshold:     r.Threshold,
		specificItems: specificItems,
		durationInSec: r.DurationInSec,
		metric:        metric,
	}
}

func newBaseTrafficShapingController(r *Rule) *baseTrafficShapingController {
	var size = 0
	if r.ParamsMaxCapacity > 0 {
		size = int(r.ParamsMaxCapacity)
	} else {
		size = int(math.Min(float64(ParamsMaxCapacity), float64(ParamsCapacityBase*r.DurationInSec)))
	}
	if size <= 0 {
		logging.Warnf("The size of cache is not more than 0, ParamsMaxCapacity: %d, ParamsCapacityBase: %d", ParamsMaxCapacity, ParamsCapacityBase)
		size = ParamsMaxCapacity
	}
	metric := &ParamsMetric{
		RuleTimeCounter:    cache.NewLRUCacheMap(size),
		RuleTokenCounter:   cache.NewLRUCacheMap(size),
		ConcurrencyCounter: cache.NewLRUCacheMap(ConcurrencyMaxCount),
	}
	return newBaseTrafficShapingControllerWithMetric(r, metric)
}

func (c *baseTrafficShapingController) BoundMetric() *ParamsMetric {
	return c.metric
}

func (c *baseTrafficShapingController) performCheckingForConcurrencyMetric(arg interface{}) *base.TokenResult {
	specificItem := c.specificItems
	initConcurrency := new(int64)
	*initConcurrency = 0
	concurrencyPtr := c.metric.ConcurrencyCounter.AddIfAbsent(arg, initConcurrency)
	if concurrencyPtr == nil {
		// First to access this arg
		return nil
	}
	concurrency := atomic.LoadInt64(concurrencyPtr)
	concurrency++
	if specificConcurrency, existed := specificItem[arg]; existed {
		if concurrency <= specificConcurrency {
			return nil
		}
		return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
			fmt.Sprintf("arg=%v", arg), c.BoundRule(), concurrency)
	}
	threshold := int64(c.threshold)
	if concurrency <= threshold {
		return nil
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
		fmt.Sprintf("arg=%v", arg), c.BoundRule(), concurrency)
}

// rejectTrafficShapingController use Reject strategy
type rejectTrafficShapingController struct {
	baseTrafficShapingController
	burstCount int64
}

// rejectTrafficShapingController use Throttling strategy
type throttlingTrafficShapingController struct {
	baseTrafficShapingController
	maxQueueingTimeMs int64
}

func (c *baseTrafficShapingController) BoundRule() *Rule {
	return c.r
}

func (c *baseTrafficShapingController) BoundParamIndex() int {
	return c.paramIndex
}

func (c *rejectTrafficShapingController) PerformChecking(arg interface{}, acquireCount int64) *base.TokenResult {
	metric := c.metric
	if metric == nil {
		return nil
	}

	if c.metricType == Concurrency {
		return c.performCheckingForConcurrencyMetric(arg)
	} else if c.metricType > QPS {
		return nil
	}

	timeCounter := metric.RuleTimeCounter
	tokenCounter := metric.RuleTokenCounter
	if timeCounter == nil || tokenCounter == nil {
		return nil
	}

	// calculate available token
	tokenCount := int64(c.threshold)
	val, existed := c.specificItems[arg]
	if existed {
		tokenCount = val
	}
	if tokenCount <= 0 {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
			fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
	}
	maxCount := tokenCount + c.burstCount
	if acquireCount > maxCount {
		// return blocked because the acquired number is more than max count of rejectTrafficShapingController
		return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
			fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
	}

	for {
		currentTimeInMs := int64(util.CurrentTimeMillis())
		lastAddTokenTimePtr := timeCounter.AddIfAbsent(arg, &currentTimeInMs)
		if lastAddTokenTimePtr == nil {
			// First to fill token, and consume token immediately
			leftCount := maxCount - acquireCount
			tokenCounter.AddIfAbsent(arg, &leftCount)
			return nil
		}

		// Calculate the time duration since last token was added.
		passTime := currentTimeInMs - atomic.LoadInt64(lastAddTokenTimePtr)
		if passTime > c.durationInSec*1000 {
			// Refill the tokens because statistic window has passed.
			leftCount := maxCount - acquireCount
			oldQpsPtr := tokenCounter.AddIfAbsent(arg, &leftCount)
			if oldQpsPtr == nil {
				// Might not be accurate here.
				atomic.StoreInt64(lastAddTokenTimePtr, currentTimeInMs)
				return nil
			} else {
				// refill token
				restQps := atomic.LoadInt64(oldQpsPtr)
				toAddTokenNum := passTime * tokenCount / (c.durationInSec * 1000)
				newQps := int64(0)
				if toAddTokenNum+restQps > maxCount {
					newQps = maxCount - acquireCount
				} else {
					newQps = toAddTokenNum + restQps - acquireCount
				}
				if newQps < 0 {
					return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
						fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
				}
				if atomic.CompareAndSwapInt64(oldQpsPtr, restQps, newQps) {
					atomic.StoreInt64(lastAddTokenTimePtr, currentTimeInMs)
					return nil
				}
				runtime.Gosched()
			}
		} else {
			//check whether the rest of token is enough to acquire
			oldQpsPtr, found := tokenCounter.Get(arg)
			if found {
				oldRestToken := atomic.LoadInt64(oldQpsPtr)
				if oldRestToken-acquireCount >= 0 {
					//update
					if atomic.CompareAndSwapInt64(oldQpsPtr, oldRestToken, oldRestToken-acquireCount) {
						return nil
					}
				} else {
					return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
						fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
				}
			}
			runtime.Gosched()
		}
	}
}

func (c *throttlingTrafficShapingController) PerformChecking(arg interface{}, acquireCount int64) *base.TokenResult {
	metric := c.metric
	if metric == nil {
		return nil
	}

	if c.metricType == Concurrency {
		return c.performCheckingForConcurrencyMetric(arg)
	} else if c.metricType > QPS {
		return nil
	}

	timeCounter := metric.RuleTimeCounter
	tokenCounter := metric.RuleTokenCounter
	if timeCounter == nil || tokenCounter == nil {
		return nil
	}

	// calculate available token
	tokenCount := int64(c.threshold)
	val, existed := c.specificItems[arg]
	if existed {
		tokenCount = val
	}
	if tokenCount <= 0 {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
			fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
	}
	intervalCostTime := int64(math.Round(float64(acquireCount * c.durationInSec * 1000 / tokenCount)))
	for {
		currentTimeInMs := int64(util.CurrentTimeMillis())
		lastPassTimePtr := timeCounter.AddIfAbsent(arg, &currentTimeInMs)
		if lastPassTimePtr == nil {
			// first access arg
			return nil
		}
		// load the last pass time
		lastPassTime := atomic.LoadInt64(lastPassTimePtr)
		// calculate the expected pass time
		expectedTime := lastPassTime + intervalCostTime

		if expectedTime <= currentTimeInMs || expectedTime-currentTimeInMs < c.maxQueueingTimeMs {
			if atomic.CompareAndSwapInt64(lastPassTimePtr, lastPassTime, currentTimeInMs) {
				awaitTime := expectedTime - currentTimeInMs
				if awaitTime > 0 {
					atomic.StoreInt64(lastPassTimePtr, expectedTime)
					return base.NewTokenResultShouldWait(uint64(awaitTime))
				}
				return nil
			} else {
				runtime.Gosched()
			}
		} else {
			return base.NewTokenResultBlockedWithCause(base.BlockTypeHotSpotParamFlow,
				fmt.Sprintf("arg=%v", arg), c.BoundRule(), nil)
		}
	}
}
