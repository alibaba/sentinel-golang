package core

import (
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"github.com/sentinel-group/sentinel-golang/core/slots/chain"
	"github.com/sentinel-group/sentinel-golang/core/slots/flow"
	"github.com/sentinel-group/sentinel-golang/core/slots/statistic"
)

type DefaultSlotChainBuilder struct {
}

func (dsc *DefaultSlotChainBuilder) Build() chain.SlotChain {
	linkedChain := chain.NewLinkedSlotChain()
	linkedChain.AddFirst(new(flow.FlowSlot))
	linkedChain.AddFirst(new(statistic.StatisticSlot))
	// add all slot
	return linkedChain
}

func NewDefaultSlotChainBuilder() *DefaultSlotChainBuilder {
	return &DefaultSlotChainBuilder{}
}

var defaultChain chain.SlotChain
var defaultNode *base.DefaultNode

func init() {
	defaultChain = NewDefaultSlotChainBuilder().Build()
	defaultNode = base.NewDefaultNode(nil)
}

func Entry(resource string) (*base.TokenResult, error) {
	resourceWrap := &base.ResourceWrapper{
		ResourceName: resource,
		ResourceType: base.INBOUND,
	}

	return defaultChain.Entry(nil, resourceWrap, defaultNode, 0, false)
}

func Exit(resource string) error {
	resourceWrap := &base.ResourceWrapper{
		ResourceName: resource,
		ResourceType: base.INBOUND,
	}
	return defaultChain.Exit(nil, resourceWrap, 1)
}
