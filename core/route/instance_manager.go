package route

import "github.com/alibaba/sentinel-golang/core/route/base"

type InstanceManager interface {
	StoreInstances(instances []*base.Instance)
	GetInstances() []*base.Instance
}

type BasicInstanceManager struct {
	Instances []*base.Instance
}

func NewBasicInstanceManager() *BasicInstanceManager {
	return &BasicInstanceManager{}
}

func (b *BasicInstanceManager) StoreInstances(instances []*base.Instance) {
	b.Instances = instances
}

func (b *BasicInstanceManager) GetInstances() []*base.Instance {
	return b.Instances
}
