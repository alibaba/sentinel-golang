package circuit_breaker

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"time"
)

type CircuitBreakerSlot struct {
}

func (b *CircuitBreakerSlot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	if len(resource) == 0 {
		return base.NewTokenResultPass()
	}
	return checkPass(ctx)
}

func checkPass(ctx *base.EntryContext) *base.TokenResult {
	breakers := getResBreakers(ctx.Resource.Name())
	for _, breaker := range breakers {
		r := breaker.Check(ctx)
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
