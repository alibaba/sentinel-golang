package flow

import (
	"context"
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type FlowSlot struct {
	RuleManager *RuleManager
}

func (fs *FlowSlot) IsContinue(lastResult base.SlotResult, ctx context.Context) bool {

	if lastResult.Status == base.ResultStatusOk {
		return true
	}
	return false
}

func (fs *FlowSlot) Entry(ctx context.Context, resourceWrap *base.ResourceWrapper, node *base.DefaultNode, count uint32) base.SlotResult {
	fmt.Println("flowSlot request number is ", node.TotalRequest())
	if fs.RuleManager == nil {
		return base.SlotResult{
			Status: base.ResultStatusOk,
		}
	}
	rules := fs.RuleManager.getRuleBySource(resourceWrap.ResourceName)
	if len(rules) == 0 {
		return base.SlotResult{
			Status: base.ResultStatusOk,
		}
	}
	success := checkFlow(ctx, resourceWrap, rules, node, count)
	if success {
		return base.SlotResult{
			Status: base.ResultStatusOk,
		}
	} else {
		return base.SlotResult{
			Status: base.ResultStatusBlocked,
		}
	}
}

func (fs *FlowSlot) Exit(ctx context.Context, resourceWrap *base.ResourceWrapper, count uint32) {

}

func checkFlow(ctx context.Context, resourceWrap *base.ResourceWrapper, rules []*rule, node *base.DefaultNode, count uint32) bool {
	if rules == nil {
		return true
	}
	for _, rule := range rules {
		if !canPass(ctx, resourceWrap, rule, node, count) {
			return false
		}
	}
	return true
}

func canPass(ctx context.Context, resourceWrap *base.ResourceWrapper, rule *rule, node *base.DefaultNode, count uint32) bool {
	return rule.controller_.CanPass(ctx, node, count)
}
