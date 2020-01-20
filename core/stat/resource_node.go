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

func NewResourceNode(resourceName string, resourceType base.ResourceType) *ResourceNode {
	return &ResourceNode{
		BaseStatNode: BaseStatNode{
			goroutineNum:   0,
			sampleCount:    base.DefaultSampleCount,
			intervalInMs:   base.DefaultIntervalInMs,
			rollingCounter: sbase.NewBucketLeapArray(base.DefaultSampleCount, base.DefaultIntervalInMs),
		},
		resourceName: resourceName,
		resourceType: resourceType,
	}
}

func NewCustomResourceNode(resourceName string, resourceType base.ResourceType, sampleCount uint32, intervalInMs uint32) *ResourceNode {
	return &ResourceNode{
		BaseStatNode: BaseStatNode{
			goroutineNum:   0,
			sampleCount:    sampleCount,
			intervalInMs:   intervalInMs,
			rollingCounter: sbase.NewBucketLeapArray(sampleCount, intervalInMs),
		},
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
	return n.rollingCounter
}
