package xds

import (
	"errors"
	"fmt"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/client"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/protocol"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/resources"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/utils"
	v3configcore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Agent struct {
	xdsServerAddr string
	virtualNode   *v3configcore.Node
	xdsClient     *client.XdsClient
	cdsProtocol   *protocol.CdsProtocol
	edsProtocol   *protocol.EdsProtocol
	ldsProtocol   *protocol.LdsProtocol
	rdsProtocol   *protocol.RdsProtocol
	stopChan      chan struct{}
	updateChan    chan resources.XdsUpdateEvent

	edsInitDone atomic.Bool
	rdsInitDone atomic.Bool
	initChan    chan struct{}

	// vhs,cluster,endpoint, listener from xds
	envoyVirtualHostMap     sync.Map
	envoyClusterMap         sync.Map //Reserved, but not used
	envoyClusterEndpointMap sync.Map
	envoyListenerMap        sync.Map //Reserved, but not used

	// stop or not
	runningStatus atomic.Bool
}

func NewXdsAgent(xdsServerAddr string, node *v3configcore.Node) (*Agent, error) {
	stopChan := make(chan struct{})
	updateChan := make(chan resources.XdsUpdateEvent, 8)
	initChan := make(chan struct{}, 2)

	xdsClient, err := client.NewXdsClient(stopChan, xdsServerAddr, node)
	if err != nil {
		return nil, err
	}

	// Init protocol handler
	ldsProtocol, _ := protocol.NewLdsProtocol(stopChan, updateChan, xdsClient)
	rdsProtocol, _ := protocol.NewRdsProtocol(stopChan, updateChan, xdsClient)
	cdsProtocol, _ := protocol.NewCdsProtocol(stopChan, updateChan, xdsClient)
	edsProtocol, _ := protocol.NewEdsProtocol(stopChan, updateChan, xdsClient)

	// Add protocol listener
	xdsClient.AddListener(ldsProtocol.ProcessProtocol, "lds", client.ListenerType)
	xdsClient.AddListener(rdsProtocol.ProcessProtocol, "rds", client.RouteType)
	xdsClient.AddListener(cdsProtocol.ProcessProtocol, "cds", client.ClusterType)
	xdsClient.AddListener(edsProtocol.ProcessProtocol, "eds", client.EndpointType)

	// Init pilot agent
	agent := &Agent{
		xdsServerAddr: xdsServerAddr,
		virtualNode:   node,
		xdsClient:     xdsClient,
		stopChan:      stopChan,
		updateChan:    updateChan,
		ldsProtocol:   ldsProtocol,
		rdsProtocol:   rdsProtocol,
		edsProtocol:   edsProtocol,
		cdsProtocol:   cdsProtocol,
		initChan:      initChan,
	}
	agent.runningStatus.Store(false)
	// Start xds/sds and wait
	if err = agent.run(); err != nil {
		return agent, err
	}

	// wait for lds/rds/cds/eds init done for the first time
	if err = agent.waitInitDone(time.Now()); err != nil {
		return agent, err
	}

	return agent, nil
}

func (a *Agent) waitInitDone(startTime time.Time) error {
	var initDoneCount int
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-a.initChan:
			initDoneCount++
			if initDoneCount == 2 {
				logging.Info("[Agent.waitInitDone] xds data init done", "cost", time.Since(startTime).String())
				return nil
			}
		case <-timer.C:
			return errors.New("wait for xds data init error")
		}
	}
}

func (a *Agent) run() error {
	if runningStatus := a.runningStatus.Load(); runningStatus {
		return nil
	}
	if err := a.xdsClient.InitXds(); err != nil {
		return err
	}

	go a.startUpdateEventLoop()
	return nil
}

func (a *Agent) startUpdateEventLoop() {
	for {
		select {
		case <-a.stopChan:
			a.Stop()
			return
		case event, ok := <-a.updateChan:
			if !ok {
				continue
			}

			switch event.Type {
			case resources.XdsEventUpdateLDS:
				//TODO:解析lds filters
				continue
			case resources.XdsEventUpdateCDS:
				continue

			case resources.XdsEventUpdateEDS:
				if xdsClusterEndpoints, ok := event.Object.([]resources.XdsClusterEndpoint); ok {
					for _, xdsClusterEndpoint := range xdsClusterEndpoints {
						a.envoyClusterEndpointMap.Store(xdsClusterEndpoint.Name, xdsClusterEndpoint)
					}
				}

				if a.edsInitDone.CompareAndSwap(false, true) {
					a.initChan <- struct{}{}
				}

			case resources.XdsEventUpdateRDS:
				if xdsRouteConfigurations, ok := event.Object.([]resources.XdsRouteConfig); ok {
					for _, xdsRouteConfiguration := range xdsRouteConfigurations {
						for _, xdsVirtualHost := range xdsRouteConfiguration.VirtualHosts {
							a.envoyVirtualHostMap.Store(xdsVirtualHost.Name, xdsVirtualHost)
						}
					}
				}

				if a.rdsInitDone.CompareAndSwap(false, true) {
					a.initChan <- struct{}{}
				}
			}
		}
	}
}

func (a *Agent) Stop() {
	if runningStatus := a.runningStatus.Load(); runningStatus {
		// make sure stop once
		a.runningStatus.Store(false)
		close(a.stopChan)
		close(a.updateChan)
		a.xdsClient.Stop()
	}
}

func convertToFullSvcName(rawHost string) (string, error) {
	hostParts := strings.Split(rawHost, ".")
	switch len(hostParts) {
	case 0:
		return "", fmt.Errorf("hostname is invalid: %s", rawHost)
	case 1: // service_name
		defaultNamespace, defaultDomain := os.Getenv(utils.EnvNamespace), os.Getenv(utils.EnvClusterDomain)
		if defaultNamespace == "" {
			return "", fmt.Errorf("ENV_NAMESPACE is empty")
		}
		if defaultDomain == "" {
			defaultDomain = utils.DefaultClusterDomain
		}

		return fmt.Sprintf("%s.%s.svc.%s", hostParts[0], defaultNamespace, defaultDomain), nil
	case 2: // service_name.namespace
		defaultDomain := os.Getenv(utils.EnvClusterDomain)
		if defaultDomain == "" {
			defaultDomain = utils.DefaultClusterDomain
		}

		return fmt.Sprintf("%s.%s.svc.%s", hostParts[0], hostParts[1], defaultDomain), nil
	case 3: // service_name.namespace.svc
		if hostParts[2] != "svc" {
			return "", fmt.Errorf("hostname is invalid: %s", rawHost)
		}

		defaultDomain := os.Getenv(utils.EnvClusterDomain)
		if defaultDomain == "" {
			defaultDomain = utils.DefaultClusterDomain
		}

		return fmt.Sprintf("%s.%s.svc.%s", hostParts[0], hostParts[1], defaultDomain), nil
	default: // service_name.namespace.svc.cluster_domain
		if hostParts[2] != "svc" {
			return "", fmt.Errorf("hostname is invalid: %s", rawHost)
		}

		return rawHost, nil
	}
}

func genClusterName(host string, port string, version string) (string, error) {
	if host == "" || port == "" {
		return "", fmt.Errorf("host or port should not be empty, host: %s, port: %s", host, port)
	}

	host, err := convertToFullSvcName(host)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("outbound|%s|%s|%s", port, version, host), nil
}

func (a *Agent) getEndpointsWithClusterName(clusterName string) (*resources.XdsClusterEndpoint, bool) {
	if v, hasEds := a.envoyClusterEndpointMap.Load(clusterName); hasEds {
		if xdsEndpoints, typeRight := v.(resources.XdsClusterEndpoint); typeRight {
			return &xdsEndpoints, true
		}

		return nil, false
	}

	return nil, false
}

func (a *Agent) GetEndpointList(host string, port string, version string) (*resources.XdsClusterEndpoint, bool, error) {
	clusterName, err := genClusterName(host, port, version)
	if err != nil {
		return nil, false, err
	}

	xdsEndpoints, exist := a.getEndpointsWithClusterName(clusterName)
	return xdsEndpoints, exist, nil
}

func (a *Agent) GetEndpointListWithClusterName(clusterName string) (*resources.XdsClusterEndpoint, bool, error) {
	xdsEndpoints, exist := a.getEndpointsWithClusterName(clusterName)
	return xdsEndpoints, exist, nil
}

func genVirtualHostName(host, port string) (string, error) {
	if host == "" || port == "" {
		return "", fmt.Errorf("host or port should not be empty, host: %s, port: %s", host, port)
	}

	host, err := convertToFullSvcName(host)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port), nil
}

func (a *Agent) getRoutesWithVirtualHostName(virtualHostName string) ([]resources.XdsRoute, bool) {
	if v, hasRds := a.envoyVirtualHostMap.Load(virtualHostName); hasRds {
		if xdsVirtualHost, typeRight := v.(resources.XdsVirtualHost); typeRight {
			return xdsVirtualHost.Routes, true
		}
		return nil, false
	}
	return nil, false
}

func (a *Agent) GetMatchHttpRouteCluster(method, host, port, path string, header map[string]string) (string, bool, error) {
	virtualHostName, err := genVirtualHostName(host, port)
	if err != nil {
		return "", false, err
	}

	routes, exist := a.getRoutesWithVirtualHostName(virtualHostName)
	if !exist {
		return "", false, nil
	}
	if len(routes) == 0 {
		return "", false, nil
	}

	for _, route := range routes {
		if route.Match != nil && route.Match.MatchPath(path) && route.Match.MatchMeta(header) {
			return route.Action.Cluster, exist, nil
		}
	}

	return "", false, nil
}
