package flow

import (
	"math"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/logging"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/util"
)

type WarmUpTrafficShapingCalculator struct {
	threshold         float64
	warmUpPeriodInSec uint32
	coldFactor        uint32
	warningToken      uint64
	maxToken          uint64
	slope             float64
	storedTokens      int64
	lastFilledTime    uint64
}

func NewWarmUpTrafficShapingCalculator(rule *Rule) *WarmUpTrafficShapingCalculator {
	if rule.WarmUpColdFactor <= 1 {
		rule.WarmUpColdFactor = config.DefaultWarmUpColdFactor
		logging.Warnf("[NewWarmUpTrafficShapingCalculator] invalid WarmUpColdFactor,use default values: %d", config.DefaultWarmUpColdFactor)
	}

	warningToken := uint64((float64(rule.WarmUpPeriodSec) * rule.Count) / float64(rule.WarmUpColdFactor-1))

	maxToken := warningToken + uint64(2*float64(rule.WarmUpPeriodSec)*rule.Count/float64(1.0+rule.WarmUpColdFactor))

	slope := float64(rule.WarmUpColdFactor-1.0) / rule.Count / float64(maxToken-warningToken)

	warmUpTrafficShapingCalculator := &WarmUpTrafficShapingCalculator{
		warmUpPeriodInSec: rule.WarmUpPeriodSec,
		coldFactor:        rule.WarmUpColdFactor,
		warningToken:      warningToken,
		maxToken:          maxToken,
		slope:             slope,
		threshold:         rule.Count,
		storedTokens:      0,
		lastFilledTime:    0,
	}

	return warmUpTrafficShapingCalculator
}

func (c *WarmUpTrafficShapingCalculator) CalculateAllowedTokens(node base.StatNode, acquireCount uint32, flag int32) float64 {
	previousQps := node.GetPreviousQPS(base.MetricEventPass)
	c.syncToken(previousQps)

	restToken := atomic.LoadInt64(&c.storedTokens)
	if restToken < 0 {
		restToken = 0
	}
	if restToken >= int64(c.warningToken) {
		aboveToken := restToken - int64(c.warningToken)
		warningQps := math.Nextafter(1.0/(float64(aboveToken)*c.slope+1.0/c.threshold), math.MaxFloat64)
		return warningQps
	} else {
		return c.threshold
	}
}

func (c *WarmUpTrafficShapingCalculator) syncToken(passQps float64) {
	currentTime := util.CurrentTimeMillis()
	currentTime = currentTime - currentTime%1000

	oldLastFillTime := atomic.LoadUint64(&c.lastFilledTime)
	if currentTime <= oldLastFillTime {
		return
	}

	oldValue := atomic.LoadInt64(&c.storedTokens)
	newValue := c.coolDownTokens(currentTime, passQps)

	if atomic.CompareAndSwapInt64(&c.storedTokens, oldValue, newValue) {
		if currentValue := atomic.AddInt64(&c.storedTokens, int64(-passQps)); currentValue < 0 {
			atomic.StoreInt64(&c.storedTokens, 0)
		}
		atomic.StoreUint64(&c.lastFilledTime, currentTime)
	}
}

func (c *WarmUpTrafficShapingCalculator) coolDownTokens(currentTime uint64, passQps float64) int64 {
	oldValue := atomic.LoadInt64(&c.storedTokens)
	newValue := oldValue

	// Prerequisites for adding a token:
	// When token consumption is much lower than the warning line
	if oldValue < int64(c.warningToken) {
		newValue = int64(float64(oldValue) + (float64(currentTime)-float64(atomic.LoadUint64(&c.lastFilledTime)))*c.threshold/1000)
	} else if oldValue > int64(c.warningToken) {
		if passQps < float64(uint32(c.threshold)/c.coldFactor) {
			newValue = int64(float64(oldValue) + float64(currentTime-atomic.LoadUint64(&c.lastFilledTime))*c.threshold/1000)
		}
	}

	if newValue <= int64(c.maxToken) {
		return newValue
	} else {
		return int64(c.maxToken)
	}
}
