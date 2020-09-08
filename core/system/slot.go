package system

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
)

type AdaptiveSlot struct {
}

func (s *AdaptiveSlot) Check(ctx *base.EntryContext) *base.TokenResult {
	if ctx == nil || ctx.Resource == nil || ctx.Resource.FlowType() != base.Inbound {
		return nil
	}
	rules := GetRules()
	result := ctx.RuleCheckResult
	for _, rule := range rules {
		passed, snapshotValue := s.doCheckRule(rule)
		if passed {
			continue
		}
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeSystemFlow, rule.MetricType.String(), rule, snapshotValue)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeSystemFlow, rule.MetricType.String(), rule, snapshotValue)
		}
		return result
	}
	return result
}

func (s *AdaptiveSlot) doCheckRule(rule *Rule) (bool, float64) {
	threshold := rule.TriggerCount
	switch rule.MetricType {
	case InboundQPS:
		qps := stat.InboundNode().GetQPS(base.MetricEventPass)
		res := qps < threshold
		return res, qps
	case Concurrency:
		n := float64(stat.InboundNode().CurrentGoroutineNum())
		res := n < threshold
		return res, n
	case AvgRT:
		rt := stat.InboundNode().AvgRT()
		res := rt < threshold
		return res, rt
	case Load:
		l := CurrentLoad()
		if l > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				return false, l
			}
		}
		return true, l
	case CpuUsage:
		c := CurrentCpuUsage()
		if c > threshold {
			if rule.Strategy != BBR || !checkBbrSimple() {
				return false, c
			}
		}
		return true, c
	default:
		return true, 0
	}
}

func checkBbrSimple() bool {
	concurrency := stat.InboundNode().CurrentGoroutineNum()
	minRt := stat.InboundNode().MinRT()
	maxComplete := stat.InboundNode().GetMaxAvg(base.MetricEventComplete)
	if concurrency > 1 && float64(concurrency) > maxComplete*minRt/1000 {
		return false
	}
	return true
}
