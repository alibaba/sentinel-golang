package xds

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/xds/bootstrap"
	"testing"
)

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
