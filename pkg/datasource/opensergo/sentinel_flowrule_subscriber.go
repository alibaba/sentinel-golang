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
	"reflect"

	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/model"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SentinelFlowRuleSubscriber struct {
	opensergoRuleAggregator *OpensergoRuleAggregator
}

func NewSentinelFlowRuleSubscriber(opensergoRuleAggregator *OpensergoRuleAggregator) *SentinelFlowRuleSubscriber {
	return &SentinelFlowRuleSubscriber{
		opensergoRuleAggregator: opensergoRuleAggregator,
	}
}

func (s SentinelFlowRuleSubscriber) OnSubscribeDataUpdate(subscribeKey model.SubscribeKey, data interface{}) (bool, error) {
	messages := data.([]protoreflect.ProtoMessage)
	switch reflect.ValueOf(subscribeKey.Kind()).Interface() {
	case reflect.ValueOf(configkind.ConfigKindRefRateLimitStrategy{}).Interface():
		return s.opensergoRuleAggregator.updateRateLimitStrategy(messages)
	case reflect.ValueOf(configkind.ConfigKindRefThrottlingStrategy{}).Interface():
		return s.opensergoRuleAggregator.updateThrottlingStrategy(messages)
	default:
		return false, nil
	}
}
