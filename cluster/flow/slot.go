package flow

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

const (
	RuleCheckSlotName  = "sentinel-cluster-flow-rule-check-slot"
	RuleCheckSlotOrder = 6000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct{}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[ClusterFlowSlot Check]Nil traffic controller found", "resourceName", res)
			continue
		}
		r := canPassCheck(res, tc, ctx.Input.BatchCount)
		if r == nil {
			// nil means pass
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
				// Handle waiting action.
				util.Sleep(nanosToWait)
			}
			continue
		}
	}
	return result
}

func canPassCheck(res string, tc *TrafficShapingController, batchCount uint32) *base.TokenResult {
	return tc.DoCheck(res, batchCount)
}
