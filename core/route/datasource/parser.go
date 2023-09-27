package datasource

import (
	"github.com/alibaba/sentinel-golang/core/route/base"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"sort"
	"strings"
)

// resolveRouting parses envoy RouteConfiguration to TrafficRouters and VirtualWorkloads
func resolveRouting(configurations []routev3.RouteConfiguration) ([]*base.TrafficRouter, []*base.VirtualWorkload) {
	var routerList []*base.TrafficRouter
	virtualWorkloadMap := map[string]*base.VirtualWorkload{}
	subsetMap := map[string]map[string]*base.Subset{}
	for i := range configurations {
		conf := &configurations[i]
		virtualHosts := conf.GetVirtualHosts()
		for _, virtualHost := range virtualHosts {
			appName := strings.Split(virtualHost.GetName(), ".")[0]
			router := &base.TrafficRouter{}
			router.Host = append(router.Host, appName)

			for _, route := range virtualHost.GetRoutes() {
				dest, subsets := convertRouteAction(route.GetRoute(), appName)
				httpRoute := &base.HTTPRoute{
					Name:  route.GetName(),
					Match: []*base.HTTPMatchRequest{convertRouteMatch(route.GetMatch())},
					Route: dest,
				}
				router.Http = append(router.Http, httpRoute)

				// add VirtualWorkload
				if _, ok := virtualWorkloadMap[appName]; !ok {
					virtualWorkloadMap[appName] = &base.VirtualWorkload{
						Host: appName,
					}
				}
				for subset := range subsets {
					if _, ok := subsetMap[appName]; !ok {
						subsetMap[appName] = map[string]*base.Subset{}
					}
					if _, ok := subsetMap[appName][subset]; !ok {
						subsetMap[appName][subset] = &base.Subset{
							Name: subset,
						}
					}
				}
			}
			routerList = append(routerList, router)
		}
	}

	// build VirtualWorkload list
	var vwList []*base.VirtualWorkload
	for vw, m := range subsetMap {
		virtualWorkloadMap[vw].Subsets = make([]*base.Subset, 0, len(m))
		for _, s := range m {
			virtualWorkloadMap[vw].Subsets = append(virtualWorkloadMap[vw].Subsets, s)
		}
	}
	for _, vw := range virtualWorkloadMap {
		vwList = append(vwList, vw)
	}
	sort.Slice(vwList, func(i, j int) bool {
		return vwList[i].Host < vwList[j].Host
	})
	for _, vw := range vwList {
		sort.Slice(vw.Subsets, func(i, j int) bool {
			return vw.Subsets[i].Name < vw.Subsets[j].Name
		})
	}
	return routerList, vwList
}

func convertRouteMatch(match *routev3.RouteMatch) *base.HTTPMatchRequest {
	mr := &base.HTTPMatchRequest{
		Headers: make(map[string]*base.StringMatch),
	}
	for _, m := range match.Headers {
		mr.Headers[m.GetName()] = convertHeaderMatcher(m)
	}
	return mr
}

func convertHeaderMatcher(matcher *routev3.HeaderMatcher) *base.StringMatch {
	// supports PresentMatch, ExactMatch, PrefixMatch, RegexMatch for now
	if matcher.GetPresentMatch() {
		return &base.StringMatch{Regex: ".*"}
	}
	sm := matcher.GetStringMatch()
	if sm == nil {
		return nil
	}
	if sm.GetExact() != "" {
		return &base.StringMatch{Exact: sm.GetExact()}
	}
	if sm.GetPrefix() != "" {
		return &base.StringMatch{Prefix: sm.GetPrefix()}
	}
	if sm.GetSafeRegex() != nil && sm.GetSafeRegex().Regex != "" {
		return &base.StringMatch{Regex: sm.GetSafeRegex().Regex}
	}
	return nil
}

func convertRouteAction(dest *routev3.RouteAction, host string) ([]*base.HTTPRouteDestination, map[string]bool) {
	// supports Cluster, WeightedClusters for now
	if dest.GetCluster() != "" {
		subset := getSubset(dest.GetCluster())
		return []*base.HTTPRouteDestination{
			{
				Weight: 1,
				Destination: &base.Destination{
					Host:   host,
					Subset: subset,
				},
			},
		}, map[string]bool{subset: true}
	}
	if dest.GetWeightedClusters() != nil {
		var destList []*base.HTTPRouteDestination
		subsets := make(map[string]bool)
		for _, cluster := range dest.GetWeightedClusters().Clusters {
			subset := getSubset(cluster.GetName())
			subsets[subset] = true
			destList = append(destList, &base.HTTPRouteDestination{
				Weight: int(cluster.GetWeight().GetValue()),
				Destination: &base.Destination{
					Host:   host,
					Subset: subset,
				},
			})
		}
		return destList, subsets
	}
	return nil, nil
}

func getSubset(cluster string) string {
	version := ""
	info := strings.Split(cluster, "|")
	if len(info) >= 3 {
		version = info[2]
	}
	return version
}
