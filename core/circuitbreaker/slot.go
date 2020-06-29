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
	if !checkPass(ctx) {
		result.ResetToBlockedFrom(base.BlockTypeCircuitBreaking, "CircuitBreaking")
	}
	return result
}

func checkPass(ctx *base.EntryContext) bool {
	breakers := getResBreakers(ctx.Resource.Name())
	for _, breaker := range breakers {
		isPass := breaker.TryPass(ctx)
		if !isPass {
			return false
		}
	}
	return true
}
