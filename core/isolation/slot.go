package isolation

import (
	"fmt"

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
	if passed, rule, msg, snapshot := checkPass(ctx); !passed {
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeIsolation, msg, rule, snapshot)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeIsolation, msg, rule, snapshot)
		}
	}
	return result
}

func checkPass(ctx *base.EntryContext) (bool, *Rule, string, uint32) {
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
				msg := fmt.Sprintf("concurrency check not pass, rule id: %s, current concurrency: %d, request batch count: %d, threshold: %d",
					rule.ID, curCount, batchCount, threshold)
				return false, rule, msg, curCount
			}
		}
	}
	return true, nil, "", curCount
}
