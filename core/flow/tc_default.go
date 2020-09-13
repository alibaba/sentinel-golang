package flow

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

type DirectTrafficShapingCalculator struct {
	threshold float64
}

func NewDirectTrafficShapingCalculator(threshold float64) *DirectTrafficShapingCalculator {
	return &DirectTrafficShapingCalculator{threshold: threshold}
}

func (d *DirectTrafficShapingCalculator) CalculateAllowedTokens(base.StatNode, uint32, int32) float64 {
	return d.threshold
}

type DefaultTrafficShapingChecker struct {
	rule *Rule
}

func NewDefaultTrafficShapingChecker(rule *Rule) *DefaultTrafficShapingChecker {
	return &DefaultTrafficShapingChecker{rule: rule}
}

func (d *DefaultTrafficShapingChecker) DoCheck(node base.StatNode, acquireCount uint32, threshold float64) *base.TokenResult {
	if node == nil {
		return nil
	}
	var curCount float64
	if d.rule.MetricType == Concurrency {
		curCount = float64(node.CurrentGoroutineNum())
	} else {
		curCount = node.GetQPS(base.MetricEventPass)
	}
	if curCount+float64(acquireCount) > threshold {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, "", d.rule, curCount)
	}
	return nil
}
