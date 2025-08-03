package route

import (
	"github.com/alibaba/sentinel-golang/core/route/base"
	"math/rand"
)

type RouterFilter interface {
	Filter(instanceList []*base.Instance, context *base.TrafficContext) ([]*base.Instance, error)
}

type BasicRouterFilter struct {
}

func NewBasicRouterFilter() *BasicRouterFilter {
	return &BasicRouterFilter{}
}

func (b *BasicRouterFilter) Filter(instanceList []*base.Instance, context *base.TrafficContext) ([]*base.Instance, error) {
	if len(instanceList) == 0 {
		return instanceList, nil
	}

	routeDestinationList := getRouteDestination(context)

	targets := make([]*base.Instance, 0)
	if len(routeDestinationList) == 0 {
		return targets, nil
	}
	appName := instanceList[0].AppName
	subset := randomSelectDestination(appName, routeDestinationList, instanceList)
	if subset == "" {
		return targets, nil
	}
	return getSubsetInstances(appName, subset, instanceList), nil
}

func getRouteDestination(context *base.TrafficContext) []*base.HTTPRouteDestination {
	routerList := GetTrafficRouterList()
	for _, router := range routerList {
		routes := router.Http
		if len(routes) == 0 {
			continue
		}
		for _, route := range routes {
			if route.IsMatch(context) {
				return route.Route
			}
		}
	}
	return nil
}

func randomSelectDestination(appName string,
	routeDestination []*base.HTTPRouteDestination, instanceList []*base.Instance) string {

	totalWeight := 0
	for _, dest := range routeDestination {
		if dest.Weight < 1 {
			totalWeight += 1
		} else {
			totalWeight += dest.Weight
		}
	}
	target := rand.Intn(totalWeight + 1)
	for _, dest := range routeDestination {
		if dest.Weight < 1 {
			target -= 1
		} else {
			target -= dest.Weight
		}
		if target <= 0 {
			result := getDestination(appName, dest.Destination, instanceList)
			if result != "" {
				return result
			}
		}
	}
	return ""
}

func getDestination(appName string, destination *base.Destination, instanceList []*base.Instance) string {
	subset := destination.Subset

	for {
		result := getSubsetInstances(appName, subset, instanceList)
		newResult := make([]*base.Instance, 0)

		newResult = append(newResult, result...)
		if len(newResult) > 0 {
			return subset
		}

		// fallback
		routeDestination := destination.Fallback
		if routeDestination == nil || routeDestination.Destination == nil {
			break
		}
		subset = routeDestination.Destination.Subset
	}
	return ""
}

func getSubsetInstances(appName string, subset string, instanceList []*base.Instance) []*base.Instance {
	virtualLoads := GetVirtualWorkloadList()
	result := make([]*base.Instance, 0)
	for _, virtualWorkload := range virtualLoads {
		if virtualWorkload.Host != appName {
			continue
		}
		for _, subsetRule := range virtualWorkload.Subsets {
			if subsetRule.Name != subset {
				continue
			}
			labels := subsetRule.Labels
			for _, instance := range instanceList {
				match := true
				for key, value := range labels {
					if value != instance.Metadata[key] { // TODO
						match = false
						break
					}
				}
				if match {
					result = append(result, instance)
				}
			}
		}
	}
	return result
}
