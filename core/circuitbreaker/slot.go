package circuitbreaker

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

type Slot struct {
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
	breakers := getResBreakers(ctx.Resource.Name())
	for _, breaker := range breakers {
		passed := breaker.TryPass(ctx)
		if !passed {
			return false, breaker.BoundRule()
		}
	}
	return true, nil
}
