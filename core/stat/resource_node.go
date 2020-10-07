package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
)

type ResourceNode struct {
	BaseStatNode

	resourceName string
	resourceType base.ResourceType
}

// NewResourceNode creates a new resource node with given name and classification.
func NewResourceNode(resourceName string, resourceType base.ResourceType) *ResourceNode {
	return &ResourceNode{
		BaseStatNode: *NewBaseStatNode(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs()),
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
