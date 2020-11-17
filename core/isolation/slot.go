package isolation

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

const (
	RuleCheckSlotName  = "sentinel-core-isolation-rule-check-slot"
	RuleCheckSlotOrder = 3000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(resource) == 0 {
		return result
	}
	if passed, rule, snapshot := checkPass(ctx); !passed {
		msg := "isolation check not pass"
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeIsolation, msg, rule, snapshot)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeIsolation, msg, rule, snapshot)
		}
	}
	return result
}

func checkPass(ctx *base.EntryContext) (bool, *Rule, uint32) {
	statNode := ctx.StatNode
	batchCount := ctx.Input.BatchCount
	curCount := uint32(0)
	for _, rule := range getRulesOfResource(ctx.Resource.Name()) {
		threshold := rule.Threshold
		if rule.MetricType == Concurrency {
			if cur := statNode.CurrentConcurrency(); cur >= 0 {
				curCount = uint32(cur)
			} else {
				curCount = 0
				logging.Error(errors.New("negative concurrency"), "Negative concurrency in isolation.checkPass()", "rule", rule)
			}
			if curCount+batchCount > threshold {
				return false, rule, curCount
			}
		}
	}
	return true, nil, curCount
}
