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
	"github.com/opensergo/opensergo-go/pkg/client"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
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

func NewOpenSergoDataSource(host string, port int, namespace string, app string) (*OpenSergoDataSource, error) {
	if len(namespace) == 0 || len(app) == 0 {
		return nil, errors.New(fmt.Sprintf("invalid parameters, namespace: %s, app: %s", namespace, app))
	}
	ds := &OpenSergoDataSource{
		client:                  *client.NewOpenSergoClient(host, port),
		namespace:               namespace,
		app:                     app,
		opensergoRuleAggregator: NewOpensergoRuleAggregator(),
	}

	ds.AddPropertyHandler(datasource.NewDefaultPropertyHandler(MixedPropertyJsonArrayParser, MixedPropertyUpdater))

	return ds, nil
}

func (ds *OpenSergoDataSource) Close() {
	subscribersAll := ds.client.GetSubscriberRegistry().GetSubscribersAll()
	subscribersAll.Range(func(key, value interface{}) bool {
		ds.client.UnsubscribeConfig(key.(subscribe.SubscribeKey))
		logging.Info(fmt.Sprintf("Unsubscribing OpenSergo config for target: %v", key))
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
	ds.RegisterSubscribeInfoOfFaulttoleranceRule()
	ds.RegisterSubscribeInfoOfFlowRuleStrategy()
	return nil
}

func (ds OpenSergoDataSource) doUpdate() {
	bytes, err := ds.ReadSource()
	if err != nil {
		logging.Warn("[OpenSergo] Succeed to read source in Initialize()", "namespace", ds.namespace, "app", ds.app, "content", fmt.Sprintf(string(bytes)))
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

	logging.Info("[OpenSergo] Succeed to read source", "namespace", ds.namespace, "app", ds.app, "content", mixedRule)
	bytes, err := json.Marshal(mixedRule)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (ds *OpenSergoDataSource) RegisterSubscribeInfoOfFaulttoleranceRule() {
	// registry SubscribeInfo of FaultToleranceRule
	faultToleranceRuleSubscribeKey := subscribe.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefFaultToleranceRule{})
	faultToleranceRuleSubscribeInfo := client.NewSubscribeInfo(*faultToleranceRuleSubscribeKey)
	faulttoleranceRuleSubscriber := NewFaulttoleranceRuleSubscriber(ds.opensergoRuleAggregator)
	faultToleranceRuleSubscribeInfo.AppendSubscriber(faulttoleranceRuleSubscriber)
	// log data for test
	faultToleranceRuleSubscribeInfo.AppendSubscriber(subscribe.DefaultSubscriber{})
	ds.client.RegisterSubscribeInfo(faultToleranceRuleSubscribeInfo)
	logging.Info(fmt.Sprintf("Subscribing OpenSergo base fault-tolerance rules for target <%v, %v>", ds.namespace, ds.app))
}

func (ds *OpenSergoDataSource) RegisterSubscribeInfoOfFlowRuleStrategy() {
	// registry SubscribeInfo of RateLimitStrategy
	rateLimitStrategySubscribeKey := subscribe.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	rateLimitStrategySubscribeInfo := client.NewSubscribeInfo(*rateLimitStrategySubscribeKey)
	rateLimitStrategySubscriber := NewFlowruleStrategySubscriber(ds.opensergoRuleAggregator)
	rateLimitStrategySubscribeInfo.AppendSubscriber(rateLimitStrategySubscriber)
	// log data for test
	rateLimitStrategySubscribeInfo.AppendSubscriber(subscribe.DefaultSubscriber{})
	ds.client.RegisterSubscribeInfo(rateLimitStrategySubscribeInfo)
	logging.Info(fmt.Sprintf("Subscribing OpenSergo base rate-limit strategies for target <%v, %v>", ds.namespace, ds.app))
	// TODO register other FlowRule Strategy
}

// NOTE: unsubscribe operation does not affect existing rules in SentinelProperty.
func (ds *OpenSergoDataSource) unSubscribeFlowRuleStrategy() {
	rateLimitStrategySubscribeKey := subscribe.NewSubscribeKey(ds.namespace, ds.app, configkind.ConfigKindRefRateLimitStrategy{})
	ds.client.UnsubscribeConfig(*rateLimitStrategySubscribeKey)
}
