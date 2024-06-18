/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package protocol

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/logging"
	"sync"

	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/client"
	route "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/config/route/v3"
	v3discovery "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/service/discovery/v3"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/resources"
	"github.com/dubbogo/gost/log/logger"
	"github.com/golang/protobuf/ptypes"
)

type RdsProtocol struct {
	xdsClientChannel *client.XdsClient
	resourcesMap     sync.Map
	stopChan         chan struct{}
	updateChan       chan resources.XdsUpdateEvent
}

var blackRouteNames = map[string]bool{
	"default":   true,
	"allow_any": true,
}

func NewRdsProtocol(stopChan chan struct{}, updateChan chan resources.XdsUpdateEvent, xdsClientChannel *client.XdsClient) (*RdsProtocol, error) {
	edsProtocol := &RdsProtocol{
		xdsClientChannel: xdsClientChannel,
		stopChan:         stopChan,
		updateChan:       updateChan,
	}
	return edsProtocol, nil
}

func (rds *RdsProtocol) GetTypeUrl() string {
	return client.EnvoyRoute
}

func (rds *RdsProtocol) SubscribeResource(resourceNames []string) error {
	return rds.xdsClientChannel.SendWithTypeUrlAndResourceNames(rds.GetTypeUrl(), resourceNames)
}

func (rds *RdsProtocol) ProcessProtocol(resp *v3discovery.DiscoveryResponse, xdsClientChannel *client.XdsClient) error {
	if resp.GetTypeUrl() != rds.GetTypeUrl() {
		return nil
	}

	xdsRouteConfigurations := make([]resources.XdsRouteConfig, 0)
	resourceNames := make([]string, 0)

	for _, resource := range resp.GetResources() {
		rdsResource := &route.RouteConfiguration{}
		if err := ptypes.UnmarshalAny(resource, rdsResource); err != nil {
			logger.Errorf("[Xds Protocol] fail to extract route configuration: %v", err)
			continue
		}
		xdsRouteConfiguration := rds.parseRoute(rdsResource)
		resourceNames = append(resourceNames, xdsRouteConfiguration.Name)
		xdsRouteConfigurations = append(xdsRouteConfigurations, xdsRouteConfiguration)
		fmt.Printf("[RdsProtocol.ProcessProtocol] route config name: %v, route config value: %v\n", rdsResource.Name, rdsResource.String())
	}

	// notify update
	updateEvent := resources.XdsUpdateEvent{
		Type:   resources.XdsEventUpdateRDS,
		Object: xdsRouteConfigurations,
	}
	rds.updateChan <- updateEvent

	info := &client.ResponseInfo{
		VersionInfo:   resp.VersionInfo,
		ResponseNonce: resp.Nonce,
		ResourceNames: resourceNames,
	}
	rds.xdsClientChannel.ApiStore.Store(client.EnvoyRoute, info)
	rds.xdsClientChannel.AckResponse(resp)

	return nil
}

func (rds *RdsProtocol) parseRoute(route *route.RouteConfiguration) resources.XdsRouteConfig {
	envoyRouteConfig := resources.XdsRouteConfig{
		Name:         route.Name,
		VirtualHosts: make(map[string]resources.XdsVirtualHost, 0),
	}
	envoyRouteConfig.Name = route.Name
	for _, vh := range route.GetVirtualHosts() {
		// virtual host
		envoyVirtualHost := resources.XdsVirtualHost{
			Name:    vh.Name,
			Domains: make([]string, 0),
			Routes:  make([]resources.XdsRoute, 0),
		}
		// domains
		for _, domain := range vh.GetDomains() {
			envoyVirtualHost.Domains = append(envoyVirtualHost.Domains, domain)
		}
		// routes
		for _, vhRoute := range vh.GetRoutes() {
			if blackRouteNames[vhRoute.GetName()] {
				logging.Debug("[RdsProtocol.parseRoute] default route, will be ignored: %v", vhRoute.String())
				continue
			}

			envoyRoute := resources.XdsRoute{}
			envoyRoute.Name = vhRoute.GetName()

			// route match
			envoyRouteMatch := &resources.HTTPRouteMatch{}
			if vhRoute.GetMatch() != nil {
				envoyRouteMatch.Path = vhRoute.GetMatch().GetPath()
				envoyRouteMatch.Prefix = vhRoute.GetMatch().GetPrefix()
				if vhRoute.GetMatch().GetCaseSensitive() != nil {
					envoyRouteMatch.CaseSensitive = vhRoute.GetMatch().GetCaseSensitive().GetValue()
				}
				if vhRoute.GetMatch().GetSafeRegex() != nil {
					envoyRouteMatch.Regex = vhRoute.GetMatch().GetSafeRegex().GetRegex()
				}
				if len(vhRoute.GetMatch().GetHeaders()) > 0 {
					envoyRouteMatch.Headers = resources.BuildMatchers(vhRoute.GetMatch().GetHeaders())
				}
			}
			envoyRoute.Match = envoyRouteMatch

			// route action
			envoyRouteAction := resources.XdsRouteAction{
				Cluster:        vhRoute.GetRoute().GetCluster(),
				ClusterWeights: make([]resources.XdsClusterWeight, 0),
			}
			envoyRouteAction.Cluster = vhRoute.GetRoute().GetCluster()
			if vhRoute.GetRoute().GetWeightedClusters() != nil {
				for _, clusterWeight := range vhRoute.GetRoute().GetWeightedClusters().GetClusters() {
					envoyClusterWeight := resources.XdsClusterWeight{}
					envoyClusterWeight.Name = clusterWeight.Name
					envoyClusterWeight.Weight = clusterWeight.Weight.Value
					envoyRouteAction.ClusterWeights = append(envoyRouteAction.ClusterWeights, envoyClusterWeight)
				}
			}
			envoyRoute.Action = envoyRouteAction

			envoyVirtualHost.Routes = append(envoyVirtualHost.Routes, envoyRoute)
			fmt.Printf("[RdsProtocol.parseRoute] vh name: %v, raw route value: %v, parsed route value: %v\n",
				vh.Name, vhRoute.String(), envoyRoute)
		}
		envoyRouteConfig.VirtualHosts[envoyVirtualHost.Name] = envoyVirtualHost
		// parse more info here like ratelimit, retry policy
	}
	return envoyRouteConfig
}
