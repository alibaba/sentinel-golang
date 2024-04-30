package traffic

import (
	"context"
	"fmt"
)

var defaultPortMap = map[string]string{
	"":      "80",
	"http":  "80",
	"https": "443",
}

type Context struct {
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
	HostName   string `json:"host"`
	Port       string `json:"port"`
	Updated    bool   `json:"updated"`
	TrafficTag string `json:"trafficTag"`
}

func Route(trafficContext *Context) (*RouteDestination, error) {
	if trafficContext == nil || trafficContext.HostName == "" {
		return nil, fmt.Errorf("trafficContext is nil or HostName is empty")
	}

	// rds匹配
	routeDestination, err := routeByRDS(trafficContext)
	if err != nil {
		fmt.Printf("[routeByRDS] route by rds err: %v\n", err)
	}
	if err == nil && routeDestination != nil && routeDestination.Updated {
		return routeDestination, nil
	}

	// cds匹配
	routeDestination, err = routeByCDS(trafficContext)
	if err != nil {
		fmt.Printf("[routeByRDS] route by cds err: %v\n", err)
	}
	if err == nil && routeDestination != nil && routeDestination.Updated {
		return routeDestination, nil
	}

	return &RouteDestination{
		HostName:   trafficContext.HostName,
		Port:       trafficContext.Port,
		Updated:    false,
		TrafficTag: trafficContext.TrafficTag,
	}, nil
}

func routeByRDS(trafficContext *Context) (*RouteDestination, error) {
	method := trafficContext.Method
	hostName := trafficContext.HostName
	port := trafficContext.Port
	path := trafficContext.Path
	header := trafficContext.Header
	trafficTag := trafficContext.TrafficTag

	trafficDestination := &RouteDestination{
		HostName:   hostName,
		Port:       port,
		Updated:    false,
		TrafficTag: trafficTag,
	}

	if port == "" {
		port = defaultPortMap[trafficContext.Scheme]
	}
	newHost, newPort, newTrafficTag, exist, err := getInstanceByRds(method, hostName, port, path, header)
	if err != nil {
		fmt.Printf("[routeByRDS] get instance by rds err: %v, host: %v, port: %v\n", err, hostName, port)
		return trafficDestination, err
	}
	if !exist {
		fmt.Printf("[routeByRDS] instance not exist, host: %v, port: %v\n", hostName, port)
		return trafficDestination, nil
	}

	trafficDestination.HostName = newHost
	trafficDestination.Port = newPort
	trafficDestination.Updated = true
	trafficDestination.TrafficTag = newTrafficTag
	return trafficDestination, nil
}

func routeByCDS(trafficContext *Context) (*RouteDestination, error) {
	hostName := trafficContext.HostName
	port := trafficContext.Port
	trafficTag := trafficContext.TrafficTag

	trafficDestination := &RouteDestination{
		HostName:   hostName,
		Port:       port,
		Updated:    false,
		TrafficTag: trafficTag,
	}

	if port == "" {
		port = defaultPortMap[trafficContext.Scheme]
	}
	if trafficTag == "" {
		trafficTag = baseVersion
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
	trafficDestination.Updated = true
	return trafficDestination, nil
}
