package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

type ResourceNode struct {
	BaseStatNode

	resourceName string
	resourceType base.ResourceType
}

// NewResourceNode creates a new resource node with given name and classification.
func NewResourceNode(resourceName string, resourceType base.ResourceType) *ResourceNode {
	return &ResourceNode{
		// TODO: make this configurable
		BaseStatNode: *NewBaseStatNode(base.DefaultSampleCount, base.DefaultIntervalMs),
		resourceName: resourceName,
		resourceType: resourceType,
	}
}

func (n *ResourceNode) ResourceType() base.ResourceType {
	return n.resourceType
}

func (n *ResourceNode) ResourceName() string {
	return n.resourceName
}
