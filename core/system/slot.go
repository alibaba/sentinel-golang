package system

import (
	"fmt"

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
	threshold := rule.TriggerCount
	switch rule.MetricType {
	case InboundQPS:
		qps := stat.InboundNode().GetQPS(base.MetricEventPass)
		res := qps < threshold
		msg := ""
		if !res {
			msg = fmt.Sprintf("qps check not pass, rule id: %s, current: %.2f, threshold: %.2f", rule.ID, qps, threshold)
		}
		return res, msg, qps
	case Concurrency:
		n := float64(stat.InboundNode().CurrentConcurrency())
		res := n < threshold
		msg := ""
		if !res {
			msg = fmt.Sprintf("concurrency check not pass, rule id: %s, current: %.2f, threshold: %.2f", rule.ID, n, threshold)
		}
		return res, msg, n
	case AvgRT:
		rt := stat.InboundNode().AvgRT()
		res := rt < threshold
		msg := ""
		if !res {
			msg = fmt.Sprintf("avg rt check not pass, rule id: %s, current: %.2f, threshold: %.2f", rule.ID, rt, threshold)
		}
		return res, msg, rt
	case Load:
		l := CurrentLoad()
		if l > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				msg := fmt.Sprintf("system load check not pass, rule id: %s, current: %0.2f, threshold: %.2f", rule.ID, l, threshold)
				return false, msg, l
			}
		}
		return true, "", l
	case CpuUsage:
		c := CurrentCpuUsage()
		if c > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				msg := fmt.Sprintf("cpu usage check not pass, rule id: %s, current: %0.2f, threshold: %.2f", rule.ID, c, threshold)
				return false, msg, c
			}
		}
		return true, "", c
	default:
		msg := fmt.Sprintf("undefined metric type, pass by default, rule id: %s", rule.ID)
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
