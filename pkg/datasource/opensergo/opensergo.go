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
	ds.AddPropertyHandler(datasource.NewDefaultPropertyHandler(MixedPropertyJsonArrayParser, MixedPropertyUpdater))

	return ds, nil
}

func (ds *OpenSergoDataSource) Close() {
	ds.client.SubscriberRegistry().ForEachSubscribeKey(func(key model.SubscribeKey) bool {
		ds.client.UnsubscribeConfig(key)
		logging.Info("[OpenSergoDatasource] Unsubscribing OpenSergo config.", "SubscribeKey", key)
		return true
	})
}

// Initialize
//
// 1. Set the handler for sentinel, to update sentinel local cache when the data from opensego was changed.
//
// 2. Start the NewOpenSergoClient.
//
// 3. Resister OpenSergo Subscribers by params
func (ds *OpenSergoDataSource) Initialize() error {
	ds.opensergoRuleAggregator.setSentinelUpdateHandler(ds.doUpdate)
	ds.client.Start()

	// TODO to add datasource-params in NewOpenSergoDataSource to decide register which subscribers for datasource
	// TODO add the deciding logic in follow
	ds.SubscribeFlowRule()
	ds.SubscribeFlowRuleStrategy()
	return nil
}

func (ds *OpenSergoDataSource) doUpdate() {
	bytes, err := ds.ReadSource()
	if err != nil {
		logging.Error(err, "[OpenSergoDatasource] Error occurred when doUpdate().", "namespace", ds.namespace, "app", ds.app)
	}

	ds.Handle(bytes)
}

func (ds *OpenSergoDataSource) ReadSource() ([]byte, error) {
	// assemble updated MixedRule
	mixedRule := new(MixedRule)
	if ds.opensergoRuleAggregator.mixedRuleCache.updateFlagMap[RuleType_FlowRule] {
		mixedRule.FlowRule = ds.opensergoRuleAggregator.mixedRuleCache.FlowRule
	}
	// TODO assembler other rule-type

	logging.Info("[OpenSergoDatasource] Succeed to read source.", "namespace", ds.namespace, "app", ds.app, "result", mixedRule)
	bytes, err := json.Marshal(mixedRule)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (ds *OpenSergoDataSource) SubscribeFlowRule() {
	// Subscribe FlowRule
	faultToleranceRuleSubscribeKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefFaultToleranceRule{})
	faulttoleranceRuleSubscriber := NewFaulttoleranceRuleSubscriber(ds.opensergoRuleAggregator)
	api.WithSubscriber(faulttoleranceRuleSubscriber)
	ds.client.SubscribeConfig(*faultToleranceRuleSubscribeKey, api.WithSubscriber(faulttoleranceRuleSubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo base fault-tolerance strategies.", "namespace", ds.namespace, "app", ds.app)
}

func (ds *OpenSergoDataSource) SubscribeFlowRuleStrategy() {
	// registry SubscribeInfo of RateLimitStrategy
	rateLimitStrategySubscribeKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	rateLimitStrategySubscriber := NewFlowruleStrategySubscriber(ds.opensergoRuleAggregator)
	ds.client.SubscribeConfig(*rateLimitStrategySubscribeKey, api.WithSubscriber(rateLimitStrategySubscriber))
	logging.Info("[OpenSergoDatasource] Subscribing OpenSergo base rate-limit strategies.", "namespace", ds.namespace, "app", ds.app)
	// TODO register other FlowRule Strategy
}

// NOTE: unsubscribe operation does not affect existing rules in SentinelProperty.
func (ds *OpenSergoDataSource) unSubscribeFlowRuleStrategy() {
	rateLimitStrategySubscribeKey := model.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	ds.client.UnsubscribeConfig(*rateLimitStrategySubscribeKey)
}
