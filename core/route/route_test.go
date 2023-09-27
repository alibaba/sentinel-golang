package route

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/route/base"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var clusterManager *ClusterManager

func TestMain(m *testing.M) {
	trafficRouter := base.TrafficRouter{
		Host: []string{"test-provider"},
		Http: []*base.HTTPRoute{
			{
				Name: "test-traffic-provider-rule-basic",
				Match: []*base.HTTPMatchRequest{
					{
						Headers: map[string]*base.StringMatch{
							"test-tag": {Exact: "basic-test"},
						},
						Method: &base.StringMatch{
							Exact: "hello",
						},
					},
				},
				Route: []*base.HTTPRouteDestination{
					{
						Weight: 1,
						Destination: &base.Destination{
							Host:   "test-provider",
							Subset: "v1",
						},
					},
				},
			}, {
				Name: "test-traffic-provider-rule-fallback",
				Match: []*base.HTTPMatchRequest{
					{
						Headers: map[string]*base.StringMatch{
							"test-tag": {Exact: "fallback-test"},
						},
						Method: &base.StringMatch{
							Exact: "hello",
						},
					},
				},
				Route: []*base.HTTPRouteDestination{
					{
						Weight: 1,
						Destination: &base.Destination{
							Host:   "test-provider",
							Subset: "v4",
							Fallback: &base.HTTPRouteDestination{
								Destination: &base.Destination{
									Host:   "test-provider",
									Subset: "v3",
								},
							},
						},
					},
				},
			},
		},
	}

	virtualWorkload := base.VirtualWorkload{
		Host: "test-provider",
		Subsets: []*base.Subset{
			{
				Name: "v1",
				Labels: map[string]string{
					"instance-tag": "v1",
				},
			}, {
				Name: "v2",
				Labels: map[string]string{
					"instance-tag": "v2",
				},
			}, {
				Name: "v3",
				Labels: map[string]string{
					"instance-tag": "v3",
				},
			},
		},
	}

	SetAppName("test-consumer")
	SetTrafficRouterList([]*base.TrafficRouter{&trafficRouter})
	SetVirtualWorkloadList([]*base.VirtualWorkload{&virtualWorkload})

	clusterManager = &ClusterManager{
		InstanceManager: NewBasicInstanceManager(),
		LoadBalancer:    NewRandomLoadBalancer(),
		RouterFilterList: []RouterFilter{
			NewBasicRouterFilter(),
		},
	}

	instanceList := []*base.Instance{
		{
			AppName: "test-provider",
			Host:    "127.0.0.1",
			Port:    80081,
			Metadata: map[string]string{
				"instance-tag": "v1",
			},
		}, {
			AppName: "test-provider",
			Host:    "127.0.0.2",
			Port:    80082,
			Metadata: map[string]string{
				"instance-tag": "v2",
			},
		}, {
			AppName: "test-provider",
			Host:    "127.0.0.3",
			Port:    80083,
			Metadata: map[string]string{
				"instance-tag": "v3",
			},
		},
	}

	clusterManager.InstanceManager.StoreInstances(instanceList)

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestRouteBasic(t *testing.T) {
	context := &base.TrafficContext{
		MethodName: "hello",
		Headers: map[string]string{
			"test-tag": "basic-test",
		},
	}
	res, err := clusterManager.GetOne(context)
	if err != nil {
		fmt.Println(err)
		t.Failed()
	}
	assert.Equal(t, "127.0.0.1", res.Host)
	assert.Equal(t, 80081, res.Port)
	assert.Equal(t, "v1", res.Metadata["instance-tag"])
}

func TestRouteFallback(t *testing.T) {
	context := &base.TrafficContext{
		MethodName: "hello",
		Headers: map[string]string{
			"test-tag": "fallback-test",
		},
	}
	res, err := clusterManager.GetOne(context)
	if err != nil {
		fmt.Println(err)
		t.Failed()
	}
	assert.Equal(t, "127.0.0.3", res.Host)
	assert.Equal(t, 80083, res.Port)
	assert.Equal(t, "v3", res.Metadata["instance-tag"])
}
