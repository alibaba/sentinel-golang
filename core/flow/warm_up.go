package flow

import (
	"math"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/util"

	"github.com/alibaba/sentinel-golang/core/base"
)

type WarmUpTrafficShapingCalculator struct {
	threshold float64
}

func NewWarmUpTrafficShapingCalculator(threshold float64) *WarmUpTrafficShapingCalculator {
	return &WarmUpTrafficShapingCalculator{threshold: threshold}
}

func (d *WarmUpTrafficShapingCalculator) CalculateAllowedTokens(base.StatNode, uint32, int32) float64 {
	return d.threshold
}

type WarmUpTrafficShapingChecker struct {
	metricType        MetricType
	warmUpPeriodInSec uint32
	coldFactor        uint32
	warningToken      uint64
	maxToken          uint64
	slope             float64
	storedTokens      *uint64
	lastFilledTime    *uint64
	count             float64
}

func NewWarmUpTrafficShapingChecker(metricType MetricType, warmUpPeriodInSec, coldFactor uint32, count float64) *WarmUpTrafficShapingChecker {
	if coldFactor <= 1 {
		return nil
	}

	warningToken := uint64((float64(warmUpPeriodInSec) * count) / float64(coldFactor-1))

	maxToken := warningToken + uint64(2*float64(warmUpPeriodInSec)*count/float64(1.0+coldFactor))

	slope := float64(coldFactor-1.0) / count / float64(maxToken-warningToken)

	warmUpTrafficShapingChecker := &WarmUpTrafficShapingChecker{
		metricType:        metricType,
		warmUpPeriodInSec: warmUpPeriodInSec,
		coldFactor:        coldFactor,
		warningToken:      warningToken,
		maxToken:          maxToken,
		slope:             slope,
		count:             count,
		storedTokens:      new(uint64),
		lastFilledTime:    new(uint64),
	}

	return warmUpTrafficShapingChecker
}

func (d *WarmUpTrafficShapingChecker) DoCheck(node base.StatNode, acquireCount uint32, threshold float64) *base.TokenResult {
	if node == nil {
		return nil
	}
	curCount := node.GetQPS(base.MetricEventPass)

	previousQps := node.GetPreviousQPS(base.MetricEventPass)

	d.syncToken(previousQps)

	restToken := atomic.LoadUint64(d.storedTokens)
	if restToken >= d.warningToken {
		aboveToken := restToken - d.warningToken
		warningQps := math.Nextafter(1.0/(float64(aboveToken)*d.slope+1.0/threshold), math.MaxFloat64)
		if curCount+float64(acquireCount) <= warningQps {
			return nil
		}
	} else {
		if curCount+float64(acquireCount) <= threshold {
			return nil
		}
	}

	return base.NewTokenResultBlocked(base.BlockTypeFlow, "Flow")
}

func (d *WarmUpTrafficShapingChecker) syncToken(passQps float64) {
	currentTime := util.CurrentTimeMillis()
	currentTime = currentTime - currentTime%1000

	oldLastFillTime := atomic.LoadUint64(d.lastFilledTime)
	if currentTime <= oldLastFillTime {
		return
	}

	oldValue := atomic.LoadUint64(d.storedTokens)
	newValue := d.coolDownTokens(currentTime, passQps)

	if atomic.CompareAndSwapUint64(d.storedTokens, oldValue, newValue) {
		if currentValue := atomic.AddUint64(d.storedTokens, uint64(0-passQps)); currentValue < 0 {
			atomic.StoreUint64(d.storedTokens, 0)
		}
		atomic.StoreUint64(d.lastFilledTime, currentTime)
	}
}

func (d *WarmUpTrafficShapingChecker) coolDownTokens(currentTime uint64, passQps float64) uint64 {
	oldValue := atomic.LoadUint64(d.storedTokens)
	newValue := oldValue

	// Prerequisites for adding a token:
	// When token consumption is much lower than the warning line
	if oldValue < d.warningToken {
		newValue = uint64(float64(oldValue) + (float64(currentTime)-float64(atomic.LoadUint64(d.lastFilledTime)))*d.count/1000)
	} else if oldValue > d.warningToken {
		if passQps < float64(uint32(d.count)/d.coldFactor) {
			newValue = uint64(float64(oldValue) + float64(currentTime-atomic.LoadUint64(d.lastFilledTime))*d.count/1000)
		}
	}

	if newValue <= d.maxToken {
		return newValue
	} else {
		return d.maxToken
	}
}
