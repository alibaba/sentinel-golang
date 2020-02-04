package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	sbase "github.com/sentinel-group/sentinel-golang/core/stat/base"
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

func (n *ResourceNode) RealBucketLeapArray() *sbase.BucketLeapArray {
	return n.arr
}
