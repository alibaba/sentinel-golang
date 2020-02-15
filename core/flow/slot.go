package flow

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/core/stat"
	"time"
)

// FlowSlot
type FlowSlot struct {
}

func (s *FlowSlot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	if len(tcs) == 0 {
		return base.NewTokenResultPass()
	}

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			logger.Warnf("nil traffic controller found, res: %s", res)
			continue
		}
		r := canPassCheck(tc, ctx.StatNode, ctx.Input.AcquireCount)
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
	return base.NewTokenResultPass()
}

func canPassCheck(tc *TrafficShapingController, node base.StatNode, acquireCount uint32) *base.TokenResult {
	return canPassCheckWithPrioritization(tc, node, acquireCount, false)
}


/** This function name "canPassCheckWithFlag" is ambiguous, and should've been renamed to "canPassCheckWithPrioritization".
	Parameter "flag" modified to "prioritized"

 */
func canPassCheckWithPrioritization(tc *TrafficShapingController, node base.StatNode, acquireCount uint32, prioritized bool) *base.TokenResult {
	if tc.rule.ClusterMode {
		return checkInCluster(tc, node, acquireCount, prioritized)
	}
	return checkInLocal(tc, node, acquireCount, prioritized)
}

func selectNodeByRelStrategy(rule *FlowRule, node base.StatNode) base.StatNode {
	if rule.RelationStrategy == AssociatedResource {
		return stat.GetResourceNode(rule.RefResource)
	}
	return node
}

func checkInLocal(tc *TrafficShapingController, node base.StatNode, acquireCount uint32, prioritized bool) *base.TokenResult {
	actual := selectNodeByRelStrategy(tc.rule, node)
	if actual == nil {
		return base.NewTokenResultPass()
	}
	return tc.PerformChecking(node, acquireCount, prioritized)
}

func checkInCluster(tc *TrafficShapingController, node base.StatNode, acquireCount uint32, prioritized bool) *base.TokenResult{
	//TODO: cluster check logic

	return fallbackToLocalOrPass(tc, node, acquireCount, prioritized)
}

func fallbackToLocalOrPass(tc *TrafficShapingController, node base.StatNode, acquireCount uint32, prioritized bool) *base.TokenResult{
	if tc.rule.ClusterConfig.FallbackToLocalWhenFail{
		return checkInLocal(tc, node, acquireCount, prioritized)
	}

	return base.NewTokenResultPass()

}
