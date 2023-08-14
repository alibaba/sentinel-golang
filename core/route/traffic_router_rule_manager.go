package route

import "github.com/alibaba/sentinel-golang/core/route/base"

type TrafficRoutingRuleGroup struct {
	AppName             string
	TrafficRouterList   []*base.TrafficRouter
	VirtualWorkloadList []*base.VirtualWorkload
}

var group = &TrafficRoutingRuleGroup{}

func SetAppName(appName string) {
	group.AppName = appName
}

func SetTrafficRouterList(list []*base.TrafficRouter) {
	group.TrafficRouterList = list
}

func SetVirtualWorkloadList(list []*base.VirtualWorkload) {
	group.VirtualWorkloadList = list
}

func GetAppName() string {
	return group.AppName
}

func GetTrafficRouterList() []*base.TrafficRouter {
	return group.TrafficRouterList
}

func GetVirtualWorkloadList() []*base.VirtualWorkload {
	return group.VirtualWorkloadList
}
