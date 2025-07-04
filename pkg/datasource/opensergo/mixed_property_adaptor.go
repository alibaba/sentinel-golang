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

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/ext/datasource"
)

type MixedRule struct {
	FlowRule             []flow.Rule
	HotSpotParamFlowRule []hotspot.Rule
	CircuitBreakerRule   []circuitbreaker.Rule
	SystemRule           []system.Rule
	IsolationRule        []isolation.Rule
}

// MixedPropertyJsonArrayParser provide JSON  as the default serialization for MixedRule
func MixedPropertyJsonArrayParser(src []byte) (interface{}, error) {
	mixedRule := new(MixedRule)
	if err := json.Unmarshal(src, mixedRule); err != nil {
		desc := fmt.Sprintf("[OpenSergoDatasource] Fail to convert source bytes to *opensergo.MixedRule, err: %s", err.Error())
		return nil, datasource.NewError(datasource.ConvertSourceError, desc)
	}
	return mixedRule, nil
}

// MixedPropertyUpdater load the newest MixedRule to downstream flow component.
func MixedPropertyUpdater(data interface{}) error {
	mixedRule := data.(*MixedRule)

	var errs []error
	flowRules := mixedRule.FlowRule
	if flowRules != nil {
		if err := datasource.FlowRulesUpdater(flowRules); err != nil {
			errs = append(errs, err)
		}
	}

	hotSpotParamFlowRule := mixedRule.HotSpotParamFlowRule
	if hotSpotParamFlowRule != nil {
		if err := datasource.HotSpotParamRulesUpdater(hotSpotParamFlowRule); err != nil {
			errs = append(errs, err)
		}
	}

	circuitBreakerRule := mixedRule.CircuitBreakerRule
	if circuitBreakerRule != nil {
		if err := datasource.CircuitBreakerRulesUpdater(circuitBreakerRule); err != nil {
			errs = append(errs, err)
		}
	}

	systemRules := mixedRule.SystemRule
	if systemRules != nil {
		if err := datasource.SystemRulesUpdater(systemRules); err != nil {
			errs = append(errs, err)
		}
	}

	isolationRule := mixedRule.IsolationRule
	if isolationRule != nil {
		if err := datasource.IsolationRulesUpdater(isolationRule); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	var errStr string
	for _, err := range errs {
		errStr = fmt.Sprintf(" | ") + fmt.Sprintf("%+v", err)
	}
	return datasource.NewError(
		datasource.UpdatePropertyError,
		errStr,
	)
}
