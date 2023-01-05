// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opensergo

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/opensergo/opensergo-go/pkg/api"
	"github.com/opensergo/opensergo-go/pkg/client"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/model"
	"github.com/pkg/errors"
)

type OpenSergoDataSource struct {
	datasource.Base
	isInitialized           util.AtomicBool
	client                  client.OpenSergoClient
	namespace               string
	app                     string
	opensergoRuleAggregator *OpensergoRuleAggregator
}

func NewOpenSergoDataSource(host string, port uint32, namespace string, app string) (*OpenSergoDataSource, error) {
	if len(namespace) == 0 || len(app) == 0 {
		return nil, errors.New(fmt.Sprintf("[OpenSergoDatasource] invalid parameters, namespace: %s, app: %s", namespace, app))
	}
	openSergoClient, err := client.NewOpenSergoClient(host, port)
	if err != nil {
		logging.Error(err, "[OpenSergoDatasource] cannot init openSergoClient.", "host", host, "port", port)
		return nil, err
	}

	ds := &OpenSergoDataSource{
		client:                  *openSergoClient,
		namespace:               namespace,
		app:                     app,
		opensergoRuleAggregator: NewOpensergoRuleAggregator(),
	}
	// add mixedRule PropertyHandler for OpenSergoDatasource
	ds.AddPropertyHandler(datasource.NewDefaultPropertyHandler(MixedPropertyJsonArrayParser, MixedPropertyUpdater))

	return ds, nil
}

func (ds *OpenSergoDataSource) Close() error {
	ds.client.SubscriberRegistry().ForEachSubscribeKey(func(key model.SubscribeKey) bool {
		ds.client.UnsubscribeConfig(key)
		logging.Info("[OpenSergoDatasource] Unsubscribing OpenSergo config.", "SubscribeKey", key)
		return true
	})
	return nil
}

// Initialize
//
// 1. Set the handler for sentinel, to update sentinel local cache when the data from opensego was changed.
//
// 2. Start the NewOpenSergoClient.
//
// 3. subscribe data from OpenSergo
func (ds *OpenSergoDataSource) Initialize() error {
	ds.opensergoRuleAggregator.setSentinelUpdateHandler(ds.doUpdate)
	if err := ds.client.Start(); err != nil {
		// TODO handle error
		return err
	}

	ds.subscribeFaultToleranceRule()
	ds.subscribeFlowRule()
	ds.subscribeIsolationRule()
	ds.subscribeCircuitBreakerRule()
	return nil
}

func (ds *OpenSergoDataSource) doUpdate() error {
	bytes, err := ds.ReadSource()
	if err != nil {
		logging.Error(err, "[OpenSergoDatasource] Error occurred when doUpdate().", "namespace", ds.namespace, "app", ds.app)
		return err
	}

	if err := ds.Handle(bytes); err != nil {
		return err
	}

	return nil
}

// ReadSource for getting mixedRule which is updated, and is needed to load into Sentinel.
func (ds *OpenSergoDataSource) ReadSource() ([]byte, error) {
	// assemble updated MixedRule
	mixedRule := new(MixedRule)
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_FlowRule] {
		mixedRule.FlowRule = ds.opensergoRuleAggregator.mixedRuleCache.FlowRule
	}
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_CircuitBreakerRule] {
		mixedRule.CircuitBreakerRule = ds.opensergoRuleAggregator.mixedRuleCache.CircuitBreakerRule
	}
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_HotSpotParamFlowRule] {
		mixedRule.HotSpotParamFlowRule = ds.opensergoRuleAggregator.mixedRuleCache.HotSpotParamFlowRule
	}
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_SystemAdaptiveRule] {
		mixedRule.SystemRule = ds.opensergoRuleAggregator.mixedRuleCache.SystemRule
	}
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_IsolationRule] {
		mixedRule.IsolationRule = ds.opensergoRuleAggregator.mixedRuleCache.IsolationRule
	}

	logging.Info("[OpenSergoDatasource] Succeed to read source.", "namespace", ds.namespace, "app", ds.app, "result", mixedRule)
	bytes, err := json.Marshal(mixedRule)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (ds *OpenSergoDataSource) subscribeFaultToleranceRule() {
	faulttoleranceRuleSubscriber := NewFaulttoleranceRuleSubscriber(ds.opensergoRuleAggregator)
	// Subscribe FaultToleranceRule
	faultToleranceRuleSubscribeKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefFaultToleranceRule{})
	api.WithSubscriber(faulttoleranceRuleSubscriber)
	ds.client.SubscribeConfig(*faultToleranceRuleSubscribeKey, api.WithSubscriber(faulttoleranceRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo base fault-tolerance strategies.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) unsubscribeFaultToleranceRule() {
	// Un-Subscribe FaultToleranceRule
	faultToleranceRuleSubscribeKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefFaultToleranceRule{})
	ds.client.UnsubscribeConfig(*faultToleranceRuleSubscribeKey)
}

func (ds *OpenSergoDataSource) subscribeFlowRule() {
	sentinelFlowRuleSubscriber := NewSentinelFlowRuleSubscriber(ds.opensergoRuleAggregator)
	// Subscribe RateLimitStrategy
	rlStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	ds.client.SubscribeConfig(*rlStrategyKey, api.WithSubscriber(sentinelFlowRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo RateLimitStrategy.", "namespace", ds.namespace, "app", ds.app)
	// Subscribe ThrottlingStrategy
	thlStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefThrottlingStrategy{})
	ds.client.SubscribeConfig(*thlStrategyKey, api.WithSubscriber(sentinelFlowRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo ThrottlingStrategy.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) unSubscribeFlowRule() {
	// Un-Subscribe RateLimitStrategy
	rlStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	ds.client.UnsubscribeConfig(*rlStrategyKey)
	logging.Info("[OpenSergoDatasource] un-subscribing OpenSergo RateLimitStrategy.", "namespace", ds.namespace, "app", ds.app)
	// Un-Subscribe ThrottlingStrategy
	thlStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefThrottlingStrategy{})
	ds.client.UnsubscribeConfig(*thlStrategyKey)
	logging.Info("[OpenSergoDatasource] un-subscribing OpenSergo ThrottlingStrategy.", "namespace", ds.namespace, "app", ds.app)
	// Un-Subscribe ConcurrencyLimitStrategy
	clStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefConcurrencyLimitStrategy{})
	ds.client.UnsubscribeConfig(*clStrategyKey)
	logging.Info("[OpenSergoDatasource] un-subscribing OpenSergo ConcurrencyLimitStrategy.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) subscribeIsolationRule() {
	isolationRuleSubscriber := NewIsolationRuleSubscriber(ds.opensergoRuleAggregator)
	// Subscribe ConcurrencyLimitStrategy
	clStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefConcurrencyLimitStrategy{})
	ds.client.SubscribeConfig(*clStrategyKey, api.WithSubscriber(isolationRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo ConcurrencyLimitStrategy.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) unSubscribeIsolationRule() {
	// Un-Subscribe ConcurrencyLimitStrategy
	clStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefConcurrencyLimitStrategy{})
	ds.client.UnsubscribeConfig(*clStrategyKey)
	logging.Info("[OpenSergoDatasource] un-subscribing OpenSergo ConcurrencyLimitStrategy.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) subscribeCircuitBreakerRule() {
	circuitBreakerRuleSubscriber := NewCircuitBreakerRuleSubscriber(ds.opensergoRuleAggregator)
	// Subscribe CircuitBreakerStrategy
	rbStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefCircuitBreakerStrategy{})
	ds.client.SubscribeConfig(*rbStrategyKey, api.WithSubscriber(circuitBreakerRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo CircuitBreakerStrategy.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) unsubscribeCircuitBreakerRule() {
	// Un-Subscribe CircuitBreakerStrategy
	rbStrategyKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefCircuitBreakerStrategy{})
	ds.client.SubscribeConfig(*rbStrategyKey)
	logging.Info("[OpenSergoDatasource] un-subscribing OpenSergo CircuitBreakerStrategy.", "namespace", ds.namespace, "app", ds.app)
}
