package datasource

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/system"
)

// FlowRulesJsonConverter provide JSON  as the default serialization for list of flow.FlowRule
func FlowRulesJsonConverter(src []byte) (interface{}, error) {
	if len(src) == 0 {
		return nil, nil
	}
	ret := make([]*flow.FlowRule, 0)
	err := json.Unmarshal(src, &ret)
	if err != nil {
		return nil, Error{
			code: ConvertSourceError,
			desc: fmt.Sprintf("Fail to unmarshal source: %s to []flow.FlowRule, err: %+v", src, err),
		}
	}
	return ret, nil
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

// SystemRulesJsonConverter provide JSON  as the default serialization for list of system.SystemRule
func SystemRulesJsonConverter(src []byte) (interface{}, error) {
	if len(src) == 0 {
		return nil, nil
	}
	ret := make([]*system.SystemRule, 0)
	err := json.Unmarshal(src, &ret)
	if err != nil {
		return nil, Error{
			code: ConvertSourceError,
			desc: fmt.Sprintf("Fail to unmarshal source: %s to []system.SystemRule, err: %+v", src, err),
		}
	}
	return ret, nil
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
