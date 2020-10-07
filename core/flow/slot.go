package flow

import (
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

type Slot struct {
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("nil traffic controller found", "resourceName", res)
			continue
		}
		r := canPassCheck(tc, ctx.StatNode, ctx.Input.AcquireCount)
		if r == nil {
			// nil means pass
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if waitMs := r.WaitMs(); waitMs > 0 {
				// Handle waiting action.
				time.Sleep(time.Duration(waitMs) * time.Millisecond)
			}
			continue
		}
	}
	return result
}

func canPassCheck(tc *TrafficShapingController, node base.StatNode, acquireCount uint32) *base.TokenResult {
	return canPassCheckWithFlag(tc, node, acquireCount, 0)
}

func canPassCheckWithFlag(tc *TrafficShapingController, node base.StatNode, acquireCount uint32, flag int32) *base.TokenResult {
	return checkInLocal(tc, node, acquireCount, flag)
}

func selectNodeByRelStrategy(rule *Rule, node base.StatNode) base.StatNode {
	if rule.RelationStrategy == AssociatedResource {
		return stat.GetResourceNode(rule.RefResource)
	}
	return node
}

func checkInLocal(tc *TrafficShapingController, resStat base.StatNode, acquireCount uint32, flag int32) *base.TokenResult {
	actual := selectNodeByRelStrategy(tc.rule, resStat)
	if actual == nil {
		logging.FrequentErrorOnce.Do(func() {
			logging.Error(errors.Errorf("nil resource node"), "no resource node for flow rule", "rule", tc.rule)
		})
		return base.NewTokenResultPass()
	}
	return tc.PerformChecking(actual, acquireCount, flag)
}
