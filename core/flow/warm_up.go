package flow

import (
	"github.com/alibaba/sentinel-golang/core/config"
	"math"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/util"

	"github.com/alibaba/sentinel-golang/core/base"
)

type WarmUpTrafficShapingCalculator struct {
	threshold         float64
	metricType        MetricType
	warmUpPeriodInSec uint32
	coldFactor        uint32
	warningToken      uint64
	maxToken          uint64
	slope             float64
	storedTokens      *uint64
	lastFilledTime    *uint64
}

func NewWarmUpTrafficShapingCalculator(metricType MetricType, warmUpPeriodInSec, warmUpColdFactor uint32, threshold float64) *WarmUpTrafficShapingCalculator {
	if warmUpColdFactor <= 1 {
		warmUpColdFactor = config.DefaultWarmUpColdFactor
	}

	warningToken := uint64((float64(warmUpPeriodInSec) * threshold) / float64(warmUpColdFactor-1))

	maxToken := warningToken + uint64(2*float64(warmUpPeriodInSec)*threshold/float64(1.0+warmUpColdFactor))

	slope := float64(warmUpColdFactor-1.0) / threshold / float64(maxToken-warningToken)

	warmUpTrafficShapingCalculator := &WarmUpTrafficShapingCalculator{
		metricType:        metricType,
		warmUpPeriodInSec: warmUpPeriodInSec,
		coldFactor:        warmUpColdFactor,
		warningToken:      warningToken,
		maxToken:          maxToken,
		slope:             slope,
		threshold:         threshold,
		storedTokens:      new(uint64),
		lastFilledTime:    new(uint64),
	}

	return warmUpTrafficShapingCalculator
}

func (d *WarmUpTrafficShapingCalculator) CalculateAllowedTokens(node base.StatNode, acquireCount uint32, flag int32) float64 {
	previousQps := node.GetPreviousQPS(base.MetricEventPass)
	d.syncToken(previousQps)

	restToken := atomic.LoadUint64(d.storedTokens)
	if restToken >= d.warningToken {
		aboveToken := restToken - d.warningToken
		warningQps := math.Nextafter(1.0/(float64(aboveToken)*d.slope+1.0/d.threshold), math.MaxFloat64)
		return warningQps
	} else {
		return d.threshold
	}
}

func (d *WarmUpTrafficShapingCalculator) syncToken(passQps float64) {
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

func (d *WarmUpTrafficShapingCalculator) coolDownTokens(currentTime uint64, passQps float64) uint64 {
	oldValue := atomic.LoadUint64(d.storedTokens)
	newValue := oldValue

	// Prerequisites for adding a token:
	// When token consumption is much lower than the warning line
	if oldValue < d.warningToken {
		newValue = uint64(float64(oldValue) + (float64(currentTime)-float64(atomic.LoadUint64(d.lastFilledTime)))*d.threshold/1000)
	} else if oldValue > d.warningToken {
		if passQps < float64(uint32(d.threshold)/d.coldFactor) {
			newValue = uint64(float64(oldValue) + float64(currentTime-atomic.LoadUint64(d.lastFilledTime))*d.threshold/1000)
		}
	}

	if newValue <= d.maxToken {
		return newValue
	} else {
		return d.maxToken
	}
}
