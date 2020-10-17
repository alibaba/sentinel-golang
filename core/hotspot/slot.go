package hotspot

import (
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	RuleCheckSlotName = "sentinel-core-hotspot-rule-check-slot"
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

// matchArg matches the arg from args based on TrafficShapingController
// return nil if match failed.
func matchArg(tc TrafficShapingController, args []interface{}) interface{} {
	if tc == nil {
		return nil
	}
	idx := tc.BoundParamIndex()
	if idx < 0 {
		idx = len(args) + idx
	}
	if idx < 0 {
		if logging.DebugEnabled() {
			logging.Debug("[Slot matchArg] The param index of hotspot traffic shaping controller is invalid", "args", args, "paramIndex", tc.BoundParamIndex())
		}
		return nil
	}
	if idx >= len(args) {
		if logging.DebugEnabled() {
			logging.Debug("[Slot matchArg] The argument in index doesn't exist", "args", args, "paramIndex", tc.BoundParamIndex())
		}
		return nil
	}
	return args[idx]
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	args := ctx.Input.Args
	batch := int64(ctx.Input.BatchCount)

	result := ctx.RuleCheckResult
	tcs := getTrafficControllersFor(res)
	for _, tc := range tcs {
		arg := matchArg(tc, args)
		if arg == nil {
			continue
		}
		r := canPassCheck(tc, arg, batch)
		if r == nil {
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

func canPassCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return canPassLocalCheck(tc, arg, batch)
}

func canPassLocalCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return tc.PerformChecking(arg, batch)
}
