package datasource

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/route/base"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"reflect"
	"testing"
)

func TestResolveRouting(t *testing.T) {
	// Mock a sample routev3.RouteConfiguration slice
	configurations := []routev3.RouteConfiguration{{
		VirtualHosts: []*routev3.VirtualHost{
			{
				Name: "Foo",
				Routes: []*routev3.Route{
					{
						Match: &routev3.RouteMatch{
							Headers: []*routev3.HeaderMatcher{
								{
									Name: "match-key-1",
									HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
										StringMatch: &matcherv3.StringMatcher{
											MatchPattern: &matcherv3.StringMatcher_Exact{
												Exact: "match-value-1",
											},
										},
									},
								},
							},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_Cluster{
									Cluster: "Foo|Bar|Sub",
								},
							},
						},
					}, {
						Match: &routev3.RouteMatch{
							Headers: []*routev3.HeaderMatcher{
								{
									Name: "match-key-2",
									HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
										StringMatch: &matcherv3.StringMatcher{
											MatchPattern: &matcherv3.StringMatcher_Exact{
												Exact: "match-value-2",
											},
										},
									},
								},
							},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_WeightedClusters{
									WeightedClusters: &routev3.WeightedCluster{
										Clusters: []*routev3.WeightedCluster_ClusterWeight{
											{
												Name:   "Foo|Bar|Sub1",
												Weight: wrapperspb.UInt32(1),
											}, {
												Name:   "Foo|Bar|Sub2",
												Weight: wrapperspb.UInt32(2),
											}, {
												Name:   "Foo|Bar|Sub3",
												Weight: wrapperspb.UInt32(3),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},
	}

	routerList, virtualWorkloads := resolveRouting(configurations)

	expectedRouterList := []*base.TrafficRouter{
		{
			Host: []string{"Foo"},
			Http: []*base.HTTPRoute{
				{
					Match: []*base.HTTPMatchRequest{
						{
							Headers: map[string]*base.StringMatch{
								"match-key-1": {
									Exact: "match-value-1",
								},
							},
						},
					},
					Route: []*base.HTTPRouteDestination{
						{
							Weight: 1,
							Destination: &base.Destination{
								Host:   "Foo",
								Subset: "Sub",
							},
						},
					},
				}, {
					Match: []*base.HTTPMatchRequest{
						{
							Headers: map[string]*base.StringMatch{
								"match-key-2": {
									Exact: "match-value-2",
								},
							},
						},
					},
					Route: []*base.HTTPRouteDestination{
						{
							Weight: 1,
							Destination: &base.Destination{
								Host:   "Foo",
								Subset: "Sub1",
							},
						}, {
							Weight: 2,
							Destination: &base.Destination{
								Host:   "Foo",
								Subset: "Sub2",
							},
						}, {
							Weight: 3,
							Destination: &base.Destination{
								Host:   "Foo",
								Subset: "Sub3",
							},
						},
					},
				},
			},
		},
	}

	expectedVirtualWorkloads := []*base.VirtualWorkload{
		{
			Host: "Foo",
			Subsets: []*base.Subset{
				{
					Name: "Sub",
				}, {
					Name: "Sub1",
				}, {
					Name: "Sub2",
				}, {
					Name: "Sub3",
				},
			},
		},
	}

	if !reflect.DeepEqual(routerList, expectedRouterList) {
		t.Errorf("Expected routerList %v, but got %v", toJSON(expectedRouterList), toJSON(routerList))
	}
	if !reflect.DeepEqual(virtualWorkloads, expectedVirtualWorkloads) {
		t.Errorf("Expected vwList %s, but got %s", toJSON(expectedVirtualWorkloads), toJSON(virtualWorkloads))
	}
}

func toJSON(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		fmt.Println("failed to parse:", err)
	}
	return string(jsonBytes)
}
