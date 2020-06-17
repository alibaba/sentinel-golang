package datasource

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
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
			desc: fmt.Sprintf("The source is invalid json(%s).", src),
		}
	}
	return true, nil
}

// FlowRulesJsonConverter provide JSON  as the default serialization for list of flow.FlowRule
func FlowRulesJsonConverter(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*flow.FlowRule, 0)
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		flowRule := &flow.FlowRule{
			Resource:          r.Get("resource").String(),
			LimitOrigin:       r.Get("limitOrigin").String(),
			MetricType:        flow.MetricType(r.Get("metricType").Int()),
			Count:             r.Get("count").Float(),
			RelationStrategy:  flow.RelationStrategy(r.Get("relationStrategy").Int()),
			ControlBehavior:   flow.ControlBehavior(r.Get("controlBehavior").Int()),
			RefResource:       r.Get("refResource").String(),
			WarmUpPeriodSec:   uint32(r.Get("warmUpPeriodSec").Int()),
			MaxQueueingTimeMs: uint32(r.Get("maxQueueingTimeMs").Int()),
			ClusterMode:       r.Get("clusterMode").Bool(),
			ClusterConfig: flow.ClusterRuleConfig{
				ThresholdType: flow.ClusterThresholdMode(r.Get("clusterConfig.thresholdType").Int()),
			},
		}
		rules = append(rules, flowRule)
	}
	return rules, nil
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
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		systemRule := &system.SystemRule{
			MetricType:   system.MetricType(r.Get("metricType").Int()),
			TriggerCount: r.Get("triggerCount").Float(),
			Strategy:     system.AdaptiveStrategy(r.Get("strategy").Int()),
		}
		rules = append(rules, systemRule)
	}
	return rules, nil
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

	rules := make([]circuitbreaker.Rule, 0)
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		if uint64(circuitbreaker.SlowRequestRatio) == r.Get("strategy").Uint() {
			rules = append(rules, circuitbreaker.NewSlowRtRule(r.Get("resource").String(),
				uint32(r.Get("statIntervalMs").Uint()), uint32(r.Get("retryTimeoutMs").Uint()),
				r.Get("maxAllowedRt").Uint(), r.Get("minRequestAmount").Uint(), r.Get("maxSlowRequestRatio").Float()))
			continue
		}
		if uint64(circuitbreaker.ErrorRatio) == r.Get("strategy").Uint() {
			rules = append(rules, circuitbreaker.NewErrorRatioRule(r.Get("resource").String(),
				uint32(r.Get("statIntervalMs").Uint()), uint32(r.Get("retryTimeoutMs").Uint()),
				r.Get("minRequestAmount").Uint(), r.Get("threshold").Float()))
			continue
		}
		if uint64(circuitbreaker.ErrorCount) == r.Get("strategy").Uint() {
			rules = append(rules, circuitbreaker.NewErrorCountRule(r.Get("resource").String(),
				uint32(r.Get("statIntervalMs").Uint()), uint32(r.Get("retryTimeoutMs").Uint()),
				r.Get("minRequestAmount").Uint(), r.Get("threshold").Uint()))
			continue
		}
		logger.Errorf("Unknown rule message: %s", r.Str)
	}
	return rules, nil
}

// CircuitBreakerRulesUpdater load the newest []circuitbreaker.Rule to downstream circuit breaker component.
func CircuitBreakerRulesUpdater(data interface{}) error {
	if data == nil {
		return circuitbreaker.ClearRules()
	}

	var rules []circuitbreaker.Rule
	if val, ok := data.([]circuitbreaker.Rule); ok {
		rules = val
	} else {
		return Error{
			code: UpdatePropertyError,
			desc: fmt.Sprintf("Fail to type assert data to []circuitbreaker.Rule, in fact, data: %+v", data),
		}
	}
	succ, err := circuitbreaker.LoadRules(rules)
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

// FrequencyParamsRulesJsonConverter provide JSON  as the default serialization for list of hotspot.Rule
func FrequencyParamsRulesJsonConverter(src []byte) (interface{}, error) {
	if valid, err := checkSrcComplianceJson(src); !valid {
		return nil, err
	}

	rules := make([]*hotspot.Rule, 0)
	result := gjson.ParseBytes(src)
	for _, r := range result.Array() {
		freqRule := &hotspot.Rule{
			Resource:          r.Get("resource").String(),
			MetricType:        hotspot.MetricType(r.Get("metricType").Int()),
			Behavior:          hotspot.ControlBehavior(r.Get("behavior").Int()),
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
			if freqRule.SpecificItems == nil {
				freqRule.SpecificItems = make(map[hotspot.SpecificValue]int64)
			}
			freqRule.SpecificItems[sp] = spItem.Get("threshold").Int()
		}
		rules = append(rules, freqRule)
	}
	return rules, nil
}

// FrequencyParamsRulesUpdater load the newest []hotspot.Rule to downstream hotspot component.
func FrequencyParamsRulesUpdater(data interface{}) error {
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

func NewFrequencyParamsRulesHandler(converter PropertyConverter) PropertyHandler {
	return NewDefaultPropertyHandler(converter, FrequencyParamsRulesUpdater)
}
