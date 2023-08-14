package route

import (
	"github.com/alibaba/sentinel-golang/core/route/base"
	"github.com/pkg/errors"
)

type ClusterManager struct {
	InstanceManager  InstanceManager
	RouterFilterList []RouterFilter
	LoadBalancer     LoadBalancer
}

func NewClusterManager(instanceManager InstanceManager, routerFilters []RouterFilter, loadBalancer LoadBalancer) *ClusterManager {
	return &ClusterManager{
		InstanceManager:  instanceManager,
		RouterFilterList: routerFilters,
		LoadBalancer:     loadBalancer,
	}
}

func (m *ClusterManager) Route(context *base.TrafficContext) ([]*base.Instance, error) {
	instances := m.InstanceManager.GetInstances()

	var err error
	for _, routerFilter := range m.RouterFilterList {
		instances, err = routerFilter.Filter(instances, context)
		if err != nil {
			return nil, err
		}
	}
	if len(instances) == 0 {
		return nil, errors.New("no matching instances")
	}
	return instances, nil
}

func (m *ClusterManager) GetOne(context *base.TrafficContext) (*base.Instance, error) {
	instances, err := m.Route(context)
	if err != nil {
		return nil, err
	}
	if m.LoadBalancer == nil {
		return instances[0], nil
	}
	instance, err := m.LoadBalancer.Select(instances, context)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
