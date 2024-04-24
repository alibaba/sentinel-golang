package gray

import (
	"github.com/alibaba/sentinel-golang/xds"
	"github.com/alibaba/sentinel-golang/xds/resources"
	"math/rand"
	"sort"
)

func getRewriteHost(host, port, version string) (string, string, error) {
	clusterEndPoint, err := getClusterEndpoints(host, port, version)
	if err != nil {
		return host, port, err
	}
	if clusterEndPoint == nil || clusterEndPoint.EndpointNum == 0 {
		return host, port, nil
	}

	newHost, newPort := selectOneEndpoint(clusterEndPoint)
	return newHost, newPort, nil
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
		if version == "" || version == baseVersion {
			return clusterEndPoint, nil
		}

		return getClusterEndpoints(host, port, baseVersion)
	}

	return clusterEndPoint, nil
}
