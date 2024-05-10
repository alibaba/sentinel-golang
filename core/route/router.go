package route

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

var defaultPortMap = map[string]string{
	"http":  "80",
	"https": "443",
}

type TrafficContext struct {
	Ctx context.Context `json:"ctx"`

	HostName string            `json:"hostName"`
	Port     string            `json:"port"`
	Header   map[string]string `json:"header"`
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Scheme   string            `json:"scheme"`

	TrafficTag string `json:"trafficTag"`
}

type RouteDestination struct {
	Ctx        context.Context
	HostName   string `json:"host"`
	Port       string `json:"port"`
	TrafficTag string `json:"trafficTag"`

	HostUpdated bool `json:"hostUpdated"`
	TagUpdated  bool `json:"tagUpdated"`
}

type RDSRouteResult struct {
	HostName    string `json:"host"`
	Port        string `json:"port"`
	TrafficTag  string `json:"trafficTag"`
	HostUpdated bool   `json:"hostUpdated"`
	TagUpdated  bool   `json:"tagUpdated"`
}

type CDSRouteResult struct {
	HostName    string `json:"host"`
	Port        string `json:"port"`
	HostUpdated bool   `json:"hostUpdated"`
}

func (t *TrafficContext) Route() (routeDestination *RouteDestination, err error) {
	if t.HostName == "" {
		return nil, errors.New("HostName is empty")
	}

	defer func() {
		if err == nil && routeDestination != nil && routeDestination.TagUpdated {
			newCtx, err := setTrafficTag(t.Ctx, t.TrafficTag)
			if err != nil {
				fmt.Printf("[TrafficContext.Route] set traffic tag err: %v, tag: %s\n", err, t.TrafficTag)
				return
			}
			t.Ctx = newCtx
			routeDestination.Ctx = newCtx
		}
	}()

	routeDestination = &RouteDestination{}
	// 获取流量标签, 如果流量标签更新, 需要更新context
	trafficTag, tagUpdated := t.getNewestTrafficTag()
	t.TrafficTag = trafficTag
	routeDestination.TrafficTag = trafficTag
	if tagUpdated {
		routeDestination.TagUpdated = true
	}

	// rds匹配
	rdsRouteResult, err := t.routeByRDS()
	if err != nil {
		fmt.Printf("[TrafficContext.Route] route by rds err: %v, traffic context: %+v\n", err, *t)
	}
	if err == nil && rdsRouteResult != nil && rdsRouteResult.HostUpdated {
		routeDestination.HostName = rdsRouteResult.HostName
		routeDestination.Port = rdsRouteResult.Port
		routeDestination.HostUpdated = rdsRouteResult.HostUpdated
		if routeDestination.TagUpdated {
			routeDestination.TrafficTag = rdsRouteResult.TrafficTag
			routeDestination.TagUpdated = rdsRouteResult.TagUpdated
		}
		return routeDestination, nil
	}

	// cds匹配
	cdsRouteResult, err := t.routeByCDS()
	if err != nil {
		fmt.Printf("[TrafficContext.Route] route by cds err: %v, traffic context: %+v\n", err, *t)
	}
	if err == nil && cdsRouteResult != nil && cdsRouteResult.HostUpdated {
		routeDestination.HostName = cdsRouteResult.HostName
		routeDestination.Port = cdsRouteResult.Port
		routeDestination.HostUpdated = cdsRouteResult.HostUpdated
		return routeDestination, nil
	}

	return &RouteDestination{}, nil
}

func (t *TrafficContext) routeByCDS() (*CDSRouteResult, error) {
	port := t.Port
	if port == "" {
		port = getDefaultPort(t.Scheme)
	}
	trafficTag := t.TrafficTag
	if trafficTag == "" {
		trafficTag = defaultTag
	}
	newHost, newPort, exist, err := getInstanceByCds(t.HostName, port, trafficTag)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &CDSRouteResult{}, nil
	}

	return &CDSRouteResult{
		HostName:    newHost,
		Port:        newPort,
		HostUpdated: true,
	}, nil
}

func (t *TrafficContext) routeByRDS() (*RDSRouteResult, error) {
	port := t.Port
	if port == "" {
		port = getDefaultPort(t.Scheme)
	}

	newHost, newPort, newTrafficTag, exist, err := getInstanceByRds(t.Method, t.HostName, port, t.Path, t.Header)
	if err != nil {
		fmt.Printf("[routeByRDS] get instance by rds err: %v, host: %v, port: %v\n", err, t.HostName, port)
		return nil, err
	}
	if !exist {
		fmt.Printf("[routeByRDS] instance not exist, host: %v, port: %v\n", t.HostName, port)
		return &RDSRouteResult{}, nil
	}

	rdsRouteResult := &RDSRouteResult{
		HostName:    newHost,
		Port:        newPort,
		HostUpdated: true,
	}
	if newTrafficTag != t.TrafficTag {
		rdsRouteResult.TrafficTag = newTrafficTag
		rdsRouteResult.TagUpdated = true
	}

	return rdsRouteResult, nil
}

func (t *TrafficContext) getNewestTrafficTag() (string, bool) {
	// 优先判断是否存在流量标签,如果存在直接返回
	trafficTag := getTrafficTag(t.Ctx)
	if trafficTag != "" {
		return trafficTag, false
	}

	// 如果不存在流量标签,使用节点标签对流量进行打标
	podTag := getPodTag(t.Ctx)
	if podTag == "" || podTag == defaultTag {
		return "", false
	}

	return podTag, true
}

func routeByCDS(trafficTrafficContext *TrafficContext) (*RouteDestination, error) {
	hostName := trafficTrafficContext.HostName
	port := trafficTrafficContext.Port
	trafficTag := trafficTrafficContext.TrafficTag

	trafficDestination := &RouteDestination{
		HostName:    hostName,
		Port:        port,
		HostUpdated: false,
		TrafficTag:  trafficTag,
	}

	if port == "" {
		port = getDefaultPort(trafficTrafficContext.Scheme)
	}
	if trafficTag == "" {
		trafficTag = defaultTag
	}
	newHost, newPort, exist, err := getInstanceByCds(hostName, port, trafficTag)
	if err != nil {
		fmt.Printf("[routeByCDS] get instance by cds err: %v, host: %v, port: %v\n", err, hostName, port)
		return trafficDestination, err
	}
	if !exist {
		fmt.Printf("[routeByCDS] instance not exist, host: %v, port: %v\n", hostName, port)
		return trafficDestination, nil
	}

	trafficDestination.HostName = newHost
	trafficDestination.Port = newPort
	trafficDestination.HostUpdated = true
	return trafficDestination, nil
}

func getDefaultPort(scheme string) string {
	if port, ok := defaultPortMap[scheme]; ok {
		return port
	}
	return defaultPort
}
