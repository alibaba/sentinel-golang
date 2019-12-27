package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

type ResourceNode struct {
	BaseStatNode

	resourceName string
	resourceType base.ResourceType
}

func NewResourceNode(resourceName string, resourceType base.ResourceType) *ResourceNode {
	return &ResourceNode{resourceName: resourceName, resourceType: resourceType}
}

func (n *ResourceNode) ResourceType() base.ResourceType {
	return n.resourceType
}

func (n *ResourceNode) ResourceName() string {
	return n.resourceName
}
