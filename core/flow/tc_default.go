package flow

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

type DirectTrafficShapingCalculator struct {
	owner     *TrafficShapingController
	threshold float64
}

func NewDirectTrafficShapingCalculator(owner *TrafficShapingController, threshold float64) *DirectTrafficShapingCalculator {
	return &DirectTrafficShapingCalculator{
		owner:     owner,
		threshold: threshold,
	}
}

func (d *DirectTrafficShapingCalculator) CalculateAllowedTokens(uint32, int32) float64 {
	return d.threshold
}

func (d *DirectTrafficShapingCalculator) BoundOwner() *TrafficShapingController {
	return d.owner
}

type RejectTrafficShapingChecker struct {
	owner *TrafficShapingController
	rule  *Rule
}

func NewRejectTrafficShapingChecker(owner *TrafficShapingController, rule *Rule) *RejectTrafficShapingChecker {
	return &RejectTrafficShapingChecker{
		owner: owner,
		rule:  rule,
	}
}

func (d *RejectTrafficShapingChecker) BoundOwner() *TrafficShapingController {
	return d.owner
}

func (d *RejectTrafficShapingChecker) DoCheck(resStat base.StatNode, acquireCount uint32, threshold float64) *base.TokenResult {
	metricReadonlyStat := d.BoundOwner().boundStat.readOnlyMetric
	if metricReadonlyStat == nil {
		return nil
	}
	curCount := float64(metricReadonlyStat.GetSum(base.MetricEventPass))
	if curCount+float64(acquireCount) > threshold {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, "", d.rule, curCount)
	}
	return nil
}
