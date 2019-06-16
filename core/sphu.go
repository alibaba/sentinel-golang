package core

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
	"github.com/sentinel-group/sentinel-golang/core/slots/chain"
	"sync"
)

var defaultChain *chain.DefaultSlotChain
var defaultNode *base.DefaultNode
var resourceWrap *base.ResourceWrapper
var lock sync.Mutex

func Entry(resource string) error {
	lock.Lock()
	if resourceWrap == nil {
		fmt.Println("default resource chain is nil, init default chain")
		resourceWrap = &base.ResourceWrapper{
			ResourceName: resource,
			ResourceType: base.INBOUND,
		}
	}
	if defaultChain == nil {
		fmt.Println("default chain is nil, init default chain")
		defaultChain = chain.NewDefaultSlotChain()
	}
	if defaultNode == nil {
		fmt.Println("default node is nil, init default node")
		defaultNode = base.NewDefaultNode(resourceWrap)
	}
	lock.Unlock()
	defaultChain.Entry(nil, resourceWrap, defaultNode, 1)
	return nil
}

func Exit(resource string) {
	resourceWrap := &base.ResourceWrapper{
		ResourceName: resource,
		ResourceType: base.INBOUND,
	}
	defaultChain.Exit(nil, resourceWrap, 1)
}
