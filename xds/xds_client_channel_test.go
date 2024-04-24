package xds

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/xds/bootstrap"
	"github.com/alibaba/sentinel-golang/xds/resources"
	"testing"
)

func TestGetMatchHttpRouteCluster(t *testing.T) {
	vh := XdsAgent.envoyVirtualHostMap
	fmt.Printf("agent.envoyVirtualHostMap info: %v\n", vh)
	v, ok := vh.Load("grpc-server-c.default.svc.cluster.local:80")
	if ok {
		routes := v.(resources.XdsVirtualHost).Routes
		for _, route := range routes {
			fmt.Printf("grpc-server-c route name: %v, match: %+v, action: %+v\n", route.Name, route.Match, route.Action)
		}
	}

	cluster, exist, err := XdsAgent.GetMatchHttpRouteCluster("GET", "grpc-server-c", "80", "/greet", map[string]string{
		"version": "v2",
		"env":     "test",
		"other":   "other",
	})
	if err != nil {
		fmt.Printf("[TestGetMatchHttpRouteCluster] get match http route cluster err: %v\n", err)
	}
	fmt.Printf("[TestGetMatchHttpRouteCluster] exist: %v, cluster: %v\n", exist, cluster)

	cluster, exist, err = XdsAgent.GetMatchHttpRouteCluster("GET", "grpc-server-c", "80", "/greet", map[string]string{
		"version": "v1",
		"env":     "test",
		"other":   "other",
	})
	if err != nil {
		fmt.Printf("[TestGetMatchHttpRouteCluster] get match http route cluster err: %v\n", err)
	}
	fmt.Printf("[TestGetMatchHttpRouteCluster] exist: %v, cluster: %v\n", exist, cluster)

	cluster, exist, err = XdsAgent.GetMatchHttpRouteCluster("GET", "grpc-server-c", "80", "/greet", map[string]string{
		"env":   "test",
		"other": "other",
	})
	if err != nil {
		fmt.Printf("[TestGetMatchHttpRouteCluster] get match http route cluster err: %v\n", err)
	}
	fmt.Printf("[TestGetMatchHttpRouteCluster] exist: %v, cluster: %v\n", exist, cluster)
}

func TestNewXdsClient(t *testing.T) {
	node, err := bootstrap.InitNode()
	if err != nil {
		t.Error(err)
	}

	agent, err := NewXdsAgent("47.97.99.1:15010", node)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("agent.envoyClusterEndpointMap info: %v\n", agent.envoyClusterEndpointMap)
	fmt.Printf("agent.envoyVirtualHostMap info: %v\n", agent.envoyVirtualHostMap)

	endPoint, exist, err := agent.GetEndpointList("gin-server-a", "80", "gray")
	if err != nil {
		fmt.Printf("[TestNewXdsClient] get end point list err: %v\n", err)
	}
	fmt.Printf("[TestNewXdsClient] exist: %v, endpoint: %v\n", exist, endPoint)

	endPoint, exist, err = agent.GetEndpointList("gin-server-a", "80", "base")
	if err != nil {
		fmt.Printf("[TestNewXdsClient] get end point list err: %v\n", err)
	}
	fmt.Printf("[TestNewXdsClient] exist: %v, endpoint: %v\n", exist, endPoint)

	endPoint, exist, err = agent.GetEndpointList("gin-server-a", "80", "")
	if err != nil {
		fmt.Printf("[TestNewXdsClient] get end point list err: %v\n", err)
	}
	fmt.Printf("[TestNewXdsClient] exist: %v, endpoint: %v\n", exist, endPoint)

	select {}
}
