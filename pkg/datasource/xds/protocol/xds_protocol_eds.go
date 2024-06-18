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
	"strconv"
	"sync"

	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/client"
	v3core "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/config/core/v3"
	envoyendpoint "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/config/endpoint/v3"
	v3discovery "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/service/discovery/v3"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/resources"
	"github.com/dubbogo/gost/log/logger"
	"github.com/golang/protobuf/ptypes"
)

type EdsProtocol struct {
	xdsClientChannel *client.XdsClient
	resourcesMap     sync.Map
	stopChan         chan struct{}
	updateChan       chan resources.XdsUpdateEvent
}

func NewEdsProtocol(stopChan chan struct{}, updateChan chan resources.XdsUpdateEvent, xdsClientChannel *client.XdsClient) (*EdsProtocol, error) {
	edsProtocol := &EdsProtocol{
		xdsClientChannel: xdsClientChannel,
		stopChan:         stopChan,
		updateChan:       updateChan,
	}
	return edsProtocol, nil
}

func (eds *EdsProtocol) GetTypeUrl() string {
	return client.EnvoyEndpoint
}

func (eds *EdsProtocol) SubscribeResource(resourceNames []string) error {
	return eds.xdsClientChannel.SendWithTypeUrlAndResourceNames(eds.GetTypeUrl(), resourceNames)
}

func (eds *EdsProtocol) ProcessProtocol(resp *v3discovery.DiscoveryResponse, xdsClientChannel *client.XdsClient) error {
	if resp.GetTypeUrl() != eds.GetTypeUrl() {
		return nil
	}

	xdsClusterEndpoints := make([]resources.XdsClusterEndpoint, 0)

	for _, resource := range resp.GetResources() {
		edsResource := &envoyendpoint.ClusterLoadAssignment{}
		if err := ptypes.UnmarshalAny(resource, edsResource); err != nil {
			logger.Errorf("[Xds Protocol] fail to extract endpoint: %v", err)
			continue
		}
		xdsClusterEndpoint, _ := eds.parseEds(edsResource)
		xdsClusterEndpoints = append(xdsClusterEndpoints, xdsClusterEndpoint)
		fmt.Printf("[EdsProtocol.ProcessProtocol] eds clusterName: %s, eds resource: %v\n", xdsClusterEndpoint.Name, xdsClusterEndpoint)
	}

	// notify update
	updateEvent := resources.XdsUpdateEvent{
		Type:   resources.XdsEventUpdateEDS,
		Object: xdsClusterEndpoints,
	}
	eds.updateChan <- updateEvent

	info := &client.ResponseInfo{
		VersionInfo:   resp.VersionInfo,
		ResponseNonce: resp.Nonce,
		ResourceNames: eds.xdsClientChannel.ApiStore.Find(client.EnvoyEndpoint).ResourceNames,
	}
	eds.xdsClientChannel.ApiStore.Store(client.EnvoyEndpoint, info)
	eds.xdsClientChannel.AckResponse(resp)
	return nil
}

func (eds *EdsProtocol) parseEds(edsResource *envoyendpoint.ClusterLoadAssignment) (resources.XdsClusterEndpoint, error) {
	clusterName := edsResource.ClusterName
	xdsClusterEndpoint := resources.XdsClusterEndpoint{
		Name: clusterName,
	}

	var totalWeight int
	endPoints := make([]resources.XdsEndpoint, 0)
	for _, lbeps := range edsResource.GetEndpoints() {
		for _, ep := range lbeps.LbEndpoints {
			if ep.GetHealthStatus() != v3core.HealthStatus_HEALTHY {
				continue
			}

			endpoint := resources.XdsEndpoint{}
			endpoint.Address = ep.GetEndpoint().GetAddress().GetSocketAddress().GetAddress()
			port := ep.GetEndpoint().GetAddress().GetSocketAddress().GetPortValue()
			endpoint.Port = strconv.FormatUint(uint64(port), 10)
			endpoint.ClusterName = clusterName
			epWeight := int(ep.GetLoadBalancingWeight().GetValue())
			if epWeight <= 0 {
				epWeight = 1
			}
			endpoint.Weight = epWeight
			endPoints = append(endPoints, endpoint)
			xdsClusterEndpoint.EndpointNum++
			totalWeight += epWeight
			xdsClusterEndpoint.StepWeight = append(xdsClusterEndpoint.StepWeight, totalWeight)
		}
	}

	xdsClusterEndpoint.TotalWeight = totalWeight
	xdsClusterEndpoint.Endpoints = endPoints
	return xdsClusterEndpoint, nil
}
