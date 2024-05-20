package nacos_v2

import (
	"encoding/json"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	RuleTypeFlow           = "FlowRule"
	RuleTypeIsolation      = "IsolationRule"
	RuleTypeSystem         = "SystemRule"
	RuleTypeCircuitBreaker = "CircuitBreakerRule"
	RuleTypeHotPot         = "HotPotRule"
)

type RuleData struct {
	Type    string
	Content string
}

func defaultOnRuleChangeHandler(namespace, group, dataId, data string) {
	logging.Info("data received for flow rules", "data", data)

	rules := make([]*RuleData, 0)
	err := json.Unmarshal([]byte(data), &rules)
	if err != nil {
		logging.Error(err, "Failed to parse flow rules")
		return
	}

	flowRules := make([]*flow.Rule, 0)
	isolationRules := make([]*isolation.Rule, 0)
	systemRules := make([]*system.Rule, 0)
	circuitBreakerRules := make([]*circuitbreaker.Rule, 0)
	paramFlowRules := make([]*hotspot.Rule, 0)
	for _, rule := range rules {
		switch rule.Type {
		case RuleTypeFlow:
			var flowRule *flow.Rule
			err = json.Unmarshal([]byte(rule.Content), &flowRule)
			if err != nil {
				logging.Error(err, "Failed to parse flow rule")
				continue
			}
			flowRules = append(flowRules, flowRule)
		case RuleTypeIsolation:
			var isolationRule *isolation.Rule
			err = json.Unmarshal([]byte(rule.Content), &isolationRule)
			if err != nil {
				logging.Error(err, "Failed to parse isolation rule")
				continue
			}
			isolationRules = append(isolationRules, isolationRule)
		case RuleTypeSystem:
			var systemRule *system.Rule
			err = json.Unmarshal([]byte(rule.Content), &systemRule)
			if err != nil {
				logging.Error(err, "Failed to parse system rule")
				continue
			}
			systemRules = append(systemRules, systemRule)
		case RuleTypeCircuitBreaker:
			var circuitBreakerRule *circuitbreaker.Rule
			err = json.Unmarshal([]byte(rule.Content), &circuitBreakerRule)
			if err != nil {
				logging.Error(err, "Failed to parse circuit breaker rule")
				continue
			}
			circuitBreakerRules = append(circuitBreakerRules, circuitBreakerRule)
		case RuleTypeHotPot:
			var paramFlowRule *hotspot.Rule
			err = json.Unmarshal([]byte(rule.Content), &paramFlowRule)
			if err != nil {
				logging.Error(err, "Failed to parse param flow rule")
				continue
			}
			paramFlowRules = append(paramFlowRules, paramFlowRule)
		default:
			continue
		}

	}

	if len(flowRules) > 0 {
		_, err = flow.LoadRules(flowRules)
		if err != nil {
			logging.Error(err, "Failed to load flow rules")
		}
	}
	if len(isolationRules) > 0 {
		_, err = isolation.LoadRules(isolationRules)
		if err != nil {
			logging.Error(err, "Failed to load isolation rules")
		}
	}
	if len(systemRules) > 0 {
		_, err = system.LoadRules(systemRules)
		if err != nil {
			logging.Error(err, "Failed to load system rules")
		}
	}
	if len(circuitBreakerRules) > 0 {
		_, err = circuitbreaker.LoadRules(circuitBreakerRules)
		if err != nil {
			logging.Error(err, "Failed to load circuit breaker rules")
		}
	}
	if len(paramFlowRules) > 0 {
		_, err = hotspot.LoadRules(paramFlowRules)
		if err != nil {
			logging.Error(err, "Failed to load param flow rules")
		}
	}
}
