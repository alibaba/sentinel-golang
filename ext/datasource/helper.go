package datasource

import (
	"encoding/json"
	"fmt"

	cb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
)

func checkSrcComplianceJson(src []byte) (bool, error) {
	if len(src) == 0 {
		return false, nil
	}
	return true, nil
}

// FlowRuleJsonArrayParser provide JSON  as the default serialization for list of flow.FlowRule
func FlowRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*flow.FlowRule, 0)
	err := json.Unmarshal(src, &rules)
	return rules, err
}

// FlowRulesUpdater load the newest []flow.FlowRule to downstream flow component.
func FlowRulesUpdater(data interface{}) error {
	if data == nil {
		return flow.ClearRules()
	}

	rules := make([]*flow.FlowRule, 0)
	if val, ok := data.([]flow.FlowRule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*flow.FlowRule); ok {
		rules = val
	} else {
		return Error{
			code: UpdatePropertyError,
			desc: fmt.Sprintf("Fail to type assert data to []flow.FlowRule or []*flow.FlowRule, in fact, data: %+v", data),
		}
	}
	succ, err := flow.LoadRules(rules)
	if succ && err == nil {
		return nil
	}
	return Error{
		code: UpdatePropertyError,
		desc: fmt.Sprintf("%+v", err),
	}
}

func NewFlowRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, FlowRulesUpdater)
}

// SystemRuleJsonArrayParser provide JSON  as the default serialization for list of system.SystemRule
func SystemRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*system.SystemRule, 0)
	err := json.Unmarshal(src, &rules)
	return rules, err
}

// SystemRulesUpdater load the newest []system.SystemRule to downstream system component.
func SystemRulesUpdater(data interface{}) error {
	if data == nil {
		return system.ClearRules()
	}

	rules := make([]*system.SystemRule, 0)
	if val, ok := data.([]system.SystemRule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*system.SystemRule); ok {
		rules = val
	} else {
		return Error{
			code: UpdatePropertyError,
			desc: fmt.Sprintf("Fail to type assert data to []system.SystemRule or []*system.SystemRule, in fact, data: %+v", data),
		}
	}
	succ, err := system.LoadRules(rules)
	if succ && err == nil {
		return nil
	}
	return Error{
		code: UpdatePropertyError,
		desc: fmt.Sprintf("%+v", err),
	}
}

func NewSystemRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, SystemRulesUpdater)
}

func CircuitBreakerRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*cb.Rule, 0)
	err := json.Unmarshal(src, &rules)
	return rules, err
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
		return Error{
			code: UpdatePropertyError,
			desc: fmt.Sprintf("Fail to type assert data to []circuitbreaker.Rule, in fact, data: %+v", data),
		}
	}
	succ, err := cb.LoadRules(rules)
	if succ && err == nil {
		return nil
	}
	return Error{
		code: UpdatePropertyError,
		desc: fmt.Sprintf("%+v", err),
	}
}

func NewCircuitBreakerRulesHandler(converter PropertyConverter) *DefaultPropertyHandler {
	return NewDefaultPropertyHandler(converter, CircuitBreakerRulesUpdater)
}

// HotSpotParamRuleJsonArrayParser decodes list of param flow rules from JSON bytes.
func HotSpotParamRuleJsonArrayParser(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*hotspot.Rule, 0)
	err := json.Unmarshal(src, &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// HotSpotParamRulesUpdater loads the provided hot-spot param rules to downstream rule manager.
func HotSpotParamRulesUpdater(data interface{}) error {
	if data == nil {
		return hotspot.ClearRules()
	}

	rules := make([]*hotspot.Rule, 0)
	if val, ok := data.([]hotspot.Rule); ok {
		for _, v := range val {
			rules = append(rules, &v)
		}
	} else if val, ok := data.([]*hotspot.Rule); ok {
		rules = val
	} else {
		return Error{
			code: UpdatePropertyError,
			desc: fmt.Sprintf("Fail to type assert data to []hotspot.Rule or []*hotspot.Rule, in fact, data: %+v", data),
		}
	}
	succ, err := hotspot.LoadRules(rules)
	if succ && err == nil {
		return nil
	}
	return Error{
		code: UpdatePropertyError,
		desc: fmt.Sprintf("%+v", err),
	}
}

func NewHotSpotParamRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, HotSpotParamRulesUpdater)
}
