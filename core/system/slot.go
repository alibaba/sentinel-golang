package system

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
)

const (
	RuleCheckSlotName  = "sentinel-core-system-adaptive-rule-check-slot"
	RuleCheckSlotOrder = 1000
)

var (
	DefaultAdaptiveSlot = &AdaptiveSlot{}
)

type AdaptiveSlot struct {
}

func (s *AdaptiveSlot) Name() string {
	return RuleCheckSlotName
}

func (s *AdaptiveSlot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *AdaptiveSlot) Check(ctx *base.EntryContext) *base.TokenResult {
	if ctx == nil || ctx.Resource == nil || ctx.Resource.FlowType() != base.Inbound {
		return nil
	}
	rules := getRules()
	result := ctx.RuleCheckResult
	for _, rule := range rules {
		passed, msg, snapshotValue := s.doCheckRule(rule)
		if passed {
			continue
		}
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeSystemFlow, msg, rule, snapshotValue)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeSystemFlow, msg, rule, snapshotValue)
		}
		return result
	}
	return result
}

func (s *AdaptiveSlot) doCheckRule(rule *Rule) (bool, string, float64) {
	var msg string

	threshold := rule.TriggerCount
	switch rule.MetricType {
	case InboundQPS:
		qps := stat.InboundNode().GetQPS(base.MetricEventPass)
		res := qps < threshold
		if !res {
			msg = "system qps check blocked"
		}
		return res, msg, qps
	case Concurrency:
		n := float64(stat.InboundNode().CurrentConcurrency())
		res := n < threshold
		if !res {
			msg = "system concurrency check blocked"
		}
		return res, msg, n
	case AvgRT:
		rt := stat.InboundNode().AvgRT()
		res := rt < threshold
		if !res {
			msg = "system avg rt check blocked"
		}
		return res, msg, rt
	case Load:
		l := CurrentLoad()
		if l > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				msg = "system load check blocked"
				return false, msg, l
			}
		}
		return true, "", l
	case CpuUsage:
		c := CurrentCpuUsage()
		if c > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				msg = "system cpu usage check blocked"
				return false, msg, c
			}
		}
		return true, "", c
	default:
		msg = "system undefined metric type, pass by default"
		return true, msg, 0.0
	}
}

func checkBbrSimple() bool {
	concurrency := stat.InboundNode().CurrentConcurrency()
	minRt := stat.InboundNode().MinRT()
	maxComplete := stat.InboundNode().GetMaxAvg(base.MetricEventComplete)
	if concurrency > 1 && float64(concurrency) > maxComplete*minRt/1000.0 {
		return false
	}
	return true
}
