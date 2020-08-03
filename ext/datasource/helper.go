package datasource

import (
	"encoding/json"
	"fmt"

	cb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/tidwall/gjson"
)

func checkSrcComplianceJson(src []byte) (bool, error) {
	if len(src) == 0 {
		return false, nil
	}
	if !gjson.ValidBytes(src) {
		return false, Error{
			code: ConvertSourceError,
			desc: fmt.Sprintf("The source is invalid json: %s", src),
		}
	}
	return true, nil
}

// FlowRulesJsonConverter provide JSON  as the default serialization for list of flow.FlowRule
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

// SystemRulesJsonConverter provide JSON  as the default serialization for list of system.SystemRule
func SystemRulesJsonConverter(src []byte) (interface{}, error) {
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

func CircuitBreakerRulesJsonConverter(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]cb.Rule, 0)
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		if uint64(cb.SlowRequestRatio) == r.Get("strategy").Uint() {
			rules = append(rules, cb.NewRule(r.Get("resource").String(), cb.SlowRequestRatio,
				cb.WithStatIntervalMs(uint32(r.Get("statIntervalMs").Uint())),
				cb.WithRetryTimeoutMs(uint32(r.Get("retryTimeoutMs").Uint())),
				cb.WithMinRequestAmount(r.Get("minRequestAmount").Uint()),
				cb.WithMaxAllowedRtMs(r.Get("maxAllowedRt").Uint()),
				cb.WithMaxSlowRequestRatio(r.Get("maxSlowRequestRatio").Float())))
			continue
		}
		if uint64(cb.ErrorRatio) == r.Get("strategy").Uint() {
			rules = append(rules, cb.NewRule(r.Get("resource").String(), cb.ErrorRatio,
				cb.WithStatIntervalMs(uint32(r.Get("statIntervalMs").Uint())),
				cb.WithRetryTimeoutMs(uint32(r.Get("retryTimeoutMs").Uint())),
				cb.WithMinRequestAmount(r.Get("minRequestAmount").Uint()),
				cb.WithErrorRatioThreshold(r.Get("threshold").Float())))
			continue
		}
		if uint64(cb.ErrorCount) == r.Get("strategy").Uint() {
			rules = append(rules, cb.NewRule(r.Get("resource").String(), cb.ErrorCount,
				cb.WithStatIntervalMs(uint32(r.Get("statIntervalMs").Uint())),
				cb.WithRetryTimeoutMs(uint32(r.Get("retryTimeoutMs").Uint())),
				cb.WithMinRequestAmount(r.Get("minRequestAmount").Uint()),
				cb.WithErrorCountThreshold(r.Get("threshold").Uint())))
			continue
		}
		logger.Errorf("Unknown rule message: %s", r.Str)
	}
	return rules, nil
}

// CircuitBreakerRulesUpdater load the newest []cb.Rule to downstream circuit breaker component.
func CircuitBreakerRulesUpdater(data interface{}) error {
	if data == nil {
		return cb.ClearRules()
	}

	var rules []cb.Rule
	if val, ok := data.([]cb.Rule); ok {
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

// HotSpotParamRulesJsonConverter decodes list of param flow rules from JSON bytes.
func HotSpotParamRulesJsonConverter(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*hotspot.Rule, 0)
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		rule := &hotspot.Rule{
			Resource:          r.Get("resource").String(),
			MetricType:        hotspot.MetricType(r.Get("metricType").Int()),
			ControlBehavior:   hotspot.ControlBehavior(r.Get("controlBehavior").Int()),
			ParamIndex:        int(r.Get("paramIndex").Int()),
			Threshold:         r.Get("threshold").Float(),
			MaxQueueingTimeMs: r.Get("maxQueueingTimeMs").Int(),
			BurstCount:        r.Get("burstCount").Int(),
			DurationInSec:     r.Get("durationInSec").Int(),
			ParamsMaxCapacity: r.Get("paramsMaxCapacity").Int(),
			SpecificItems:     nil,
		}
		for _, spItem := range r.Get("specificItems").Array() {
			sp := hotspot.SpecificValue{
				ValKind: hotspot.ParamKind(spItem.Get("valKind").Int()),
				ValStr:  spItem.Get("valStr").String(),
			}
			if rule.SpecificItems == nil {
				rule.SpecificItems = make(map[hotspot.SpecificValue]int64)
			}
			rule.SpecificItems[sp] = spItem.Get("threshold").Int()
		}
		rules = append(rules, rule)
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
