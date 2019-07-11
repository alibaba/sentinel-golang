package flow

import (
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"github.com/sentinel-group/sentinel-golang/core/slots/chain"
)

type FlowSlot struct {
	chain.LinkedSlot
	RuleManager *RuleManager
}

func (fs *FlowSlot) Entry(ctx *base.Context, resWrapper *base.ResourceWrapper, node *base.DefaultNode, count int, prioritized bool) (*base.TokenResult, error) {
	// no rule return pass
	if fs.RuleManager == nil {
		return fs.FireEntry(ctx, resWrapper, node, count, false)
	}
	rules := fs.RuleManager.getRuleBySource(resWrapper.ResourceName)
	if len(rules) == 0 {
		return fs.FireEntry(ctx, resWrapper, node, count, false)
	}
	success := checkFlow(ctx, resWrapper, rules, node, count)
	if success {
		return fs.FireEntry(ctx, resWrapper, node, count, false)
	} else {
		return base.NewSlotResultBlock("FlowSlot"), nil
	}
}

func (fs *FlowSlot) Exit(ctx *base.Context, resourceWrapper *base.ResourceWrapper, count int) error {
	return fs.FireExit(ctx, resourceWrapper, count)
}

func checkFlow(ctx *base.Context, resourceWrap *base.ResourceWrapper, rules []*rule, node *base.DefaultNode, count int) bool {
	if rules == nil {
		return true
	}
	for _, rule := range rules {
		if !canPass(ctx, resourceWrap, rule, node, uint32(count)) {
			return false
		}
	}
	return true
}

func canPass(ctx *base.Context, resourceWrap *base.ResourceWrapper, rule *rule, node *base.DefaultNode, count uint32) bool {
	return rule.controller_.CanPass(ctx, node, count)
}
