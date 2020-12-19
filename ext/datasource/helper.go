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

package datasource

import (
	"encoding/json"
	"fmt"

	cb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/system"
)

func checkSrcComplianceJson(src []byte) (bool, error) {
	if len(src) == 0 {
		return false, nil
	}
	return true, nil
}

// FlowRuleJsonArrayParser provide JSON  as the default serialization for list of flow.Rule
func FlowRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*flow.Rule, 0, 8)
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("Fail to convert source bytes to []*flow.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// FlowRulesUpdater load the newest []flow.Rule to downstream flow component.
func FlowRulesUpdater(data interface{}) error {
	if data == nil {
		return flow.ClearRules()
	}

	rules := make([]*flow.Rule, 0, 8)
	if val, ok := data.([]flow.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*flow.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("Fail to type assert data to []flow.Rule or []*flow.Rule, in fact, data: %+v", data),
		)
	}
	_, err := flow.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewFlowRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, FlowRulesUpdater)
}

// SystemRuleJsonArrayParser provide JSON  as the default serialization for list of system.Rule
func SystemRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*system.Rule, 0, 8)
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("Fail to convert source bytes to []*system.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// SystemRulesUpdater load the newest []system.Rule to downstream system component.
func SystemRulesUpdater(data interface{}) error {
	if data == nil {
		return system.ClearRules()
	}

	rules := make([]*system.Rule, 0, 8)
	if val, ok := data.([]system.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*system.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("Fail to type assert data to []system.Rule or []*system.Rule, in fact, data: %+v", data),
		)
	}
	_, err := system.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewSystemRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, SystemRulesUpdater)
}

func CircuitBreakerRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*cb.Rule, 0, 8)
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("Fail to convert source bytes to []*circuitbreaker.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// CircuitBreakerRulesUpdater load the newest []cb.Rule to downstream circuit breaker component.
func CircuitBreakerRulesUpdater(data interface{}) error {
	if data == nil {
		return cb.ClearRules()
	}

	var rules []*cb.Rule
	if val, ok := data.([]*cb.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("Fail to type assert data to []*circuitbreaker.Rule, in fact, data: %+v", data),
		)
	}
	_, err := cb.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewCircuitBreakerRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, CircuitBreakerRulesUpdater)
}

// HotSpotParamRuleJsonArrayParser decodes list of param flow rules from JSON bytes.
func HotSpotParamRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	hotspotRules := make([]*HotspotRule, 0, 8)
	if err := json.Unmarshal(src, &hotspotRules); err != nil {
		desc := fmt.Sprintf("Fail to convert source bytes to []*hotspot.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	rules := make([]*hotspot.Rule, len(hotspotRules))
	for i, hotspotRule := range hotspotRules {
		rules[i] = &hotspot.Rule{
			ID:                hotspotRule.ID,
			Resource:          hotspotRule.Resource,
			MetricType:        hotspotRule.MetricType,
			ControlBehavior:   hotspotRule.ControlBehavior,
			ParamIndex:        hotspotRule.ParamIndex,
			Threshold:         hotspotRule.Threshold,
			MaxQueueingTimeMs: hotspotRule.MaxQueueingTimeMs,
			BurstCount:        hotspotRule.BurstCount,
			DurationInSec:     hotspotRule.DurationInSec,
			ParamsMaxCapacity: hotspotRule.ParamsMaxCapacity,
			SpecificItems:     parseSpecificItems(hotspotRule.SpecificItems),
		}
	}
	return rules, nil
}

// HotSpotParamRulesUpdater loads the provided hot-spot param rules to downstream rule manager.
func HotSpotParamRulesUpdater(data interface{}) error {
	if data == nil {
		return hotspot.ClearRules()
	}

	rules := make([]*hotspot.Rule, 0, 8)
	if val, ok := data.([]hotspot.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*hotspot.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("Fail to type assert data to []hotspot.Rule or []*hotspot.Rule, in fact, data: %+v", data),
		)
	}
	_, err := hotspot.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewHotSpotParamRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, HotSpotParamRulesUpdater)
}

// IsolationRuleJsonArrayParser provide JSON  as the default serialization for list of isolation.Rule
func IsolationRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*isolation.Rule, 0, 8)
	if err := json.Unmarshal(src, &rules); err != nil {
		desc := fmt.Sprintf("Fail to convert source bytes to []*isolation.Rule, err: %s", err.Error())
		return nil, NewError(ConvertSourceError, desc)
	}
	return rules, nil
}

// IsolationRulesUpdater load the newest []isolation.Rule to downstream system component.
func IsolationRulesUpdater(data interface{}) error {
	if data == nil {
		return isolation.ClearRules()
	}

	rules := make([]*isolation.Rule, 0, 8)
	if val, ok := data.([]isolation.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*isolation.Rule); ok {
		rules = val
	} else {
		return NewError(
			UpdatePropertyError,
			fmt.Sprintf("Fail to type assert data to []isolation.Rule or []*isolation.Rule, in fact, data: %+v", data),
		)
	}
	_, err := isolation.LoadRules(rules)
	if err == nil {
		return nil
	}
	return NewError(
		UpdatePropertyError,
		fmt.Sprintf("%+v", err),
	)
}

func NewIsolationRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, IsolationRulesUpdater)
}
