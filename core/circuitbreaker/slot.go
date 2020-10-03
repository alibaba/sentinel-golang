package circuitbreaker

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

const (
	RuleCheckSlotName = "sentinel-circuit-breaker-rule-check-slot"
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

func (b *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(resource) == 0 {
		return result
	}
	if passed, rule := checkPass(ctx); !passed {
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeCircuitBreaking, "", rule, nil)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeCircuitBreaking, "", rule, nil)
		}
	}
	return result
}

func checkPass(ctx *base.EntryContext) (bool, *Rule) {
	breakers := getBreakersOfResource(ctx.Resource.Name())
	for _, breaker := range breakers {
		passed := breaker.TryPass(ctx)
		if !passed {
			return false, breaker.BoundRule()
		}
	}
	return true, nil
}
