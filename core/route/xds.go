package route

import (
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/resources"
	"math/rand"
	"sort"
	"strings"
)

func getInstanceByRds(method, host, port, path string, header map[string]string) (string, string, string, bool, error) {
	cluster, exist, err := xds.XdsAgent.GetMatchHttpRouteCluster(method, host, port, path, header)
	if err != nil {
		return "", "", "", false, err
	}
	if !exist {
		return host, port, "", false, nil
	}

	clusterEndPoint, exist, err := xds.XdsAgent.GetEndpointListWithClusterName(cluster)
	if err != nil {
		return "", "", "", false, err
	}
	if !exist || clusterEndPoint == nil || clusterEndPoint.EndpointNum == 0 {
		return host, port, "", false, nil
	}

	newHost, newPort := selectOneEndpoint(clusterEndPoint)
	var trafficTag string
	strList := strings.Split(cluster, "|")
	if len(strList) >= 2 {
		trafficTag = strList[2]
	}

	return newHost, newPort, trafficTag, true, nil
}

func getInstanceByCds(host, port, version string) (string, string, bool, error) {
	clusterEndPoint, err := getClusterEndpoints(host, port, version)
	if err != nil {
		return host, port, false, err
	}
	if clusterEndPoint == nil || clusterEndPoint.EndpointNum == 0 {
		return host, port, false, nil
	}

	newHost, newPort := selectOneEndpoint(clusterEndPoint)
	return newHost, newPort, true, nil
}

func selectOneEndpoint(clusterEndpoint *resources.XdsClusterEndpoint) (string, string) {
	r := rand.Intn(clusterEndpoint.TotalWeight) + 1
	i := sort.SearchInts(clusterEndpoint.StepWeight, r)
	return clusterEndpoint.Endpoints[i].Address, clusterEndpoint.Endpoints[i].Port
}

func getClusterEndpoints(host, port, version string) (*resources.XdsClusterEndpoint, error) {
	clusterEndPoint, exist, err := xds.XdsAgent.GetEndpointList(host, port, version)
	if err != nil {
		return nil, err
	}

	if !exist || clusterEndPoint.EndpointNum == 0 {
		if version == "" || version == DefaultTag {
			return clusterEndPoint, nil
		}

		return getClusterEndpoints(host, port, DefaultTag)
	}

	return clusterEndPoint, nil
}
