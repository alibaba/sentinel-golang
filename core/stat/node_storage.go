package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/logging"
	"sync"
)

type ResourceNodeMap map[string]*ResourceNode

var (
	logger = logging.GetDefaultLogger()

	resNodeMap = make(ResourceNodeMap, 0)
	rnsMux     = new(sync.RWMutex)
)

func GetResourceNode(resource string) *ResourceNode {
	rnsMux.RLock()
	defer rnsMux.RUnlock()

	return resNodeMap[resource]
}

func GetOrCreateResourceNode(resource string, resourceType base.ResourceType) *ResourceNode {
	node := GetResourceNode(resource)
	if node != nil {
		return node
	}
	rnsMux.Lock()
	defer rnsMux.Unlock()

	node = resNodeMap[resource]
	if node != nil {
		return node
	}

	if len(resNodeMap) >= int(base.DefaultMaxResourceAmount) {
		logger.Warnf("Resource amount exceeds the threshold: %d.", base.DefaultMaxResourceAmount)
	}
	node = NewResourceNode(resource, resourceType)
	resNodeMap[resource] = node
	return node
}

func ResetResourceNodes() {
	rnsMux.Lock()
	defer rnsMux.Unlock()
	resNodeMap = make(ResourceNodeMap, 0)
}
