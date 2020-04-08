package circuit_breaker

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/logging"
	"strings"
	"sync"
)

var (
	logger = logging.GetDefaultLogger()
)

var (
	breakerRules = make(map[string][]Rule)
	breakers     = make(map[string][]CircuitBreaker)
	ruleMux      = &sync.RWMutex{}
)

func GetResRules(resource string) []Rule {
	ruleMux.Lock()
	ret, ok := breakerRules[resource]
	ruleMux.Unlock()
	if !ok {
		ret = make([]Rule, 0)
	}
	return ret
}

// Load the newer rules to manager.
// parameter:
// rules: the newer rules, if len of rules is 0, will clear all rules of manager.
// return value:
// bool: was designed to indicate whether the internal map has been changed
// error: was designed to indicate whether occurs the error.
func LoadRules(rules []Rule) (bool, error) {
	// TODO in order to avoid invalid update, should check consistent with last update rules
	err := onRuleUpdate(rules)
	return true, err
}

func getResBreakers(resource string) []CircuitBreaker {
	ruleMux.Lock()
	ret, ok := breakers[resource]
	ruleMux.Unlock()
	if !ok {
		ret = make([]CircuitBreaker, 0)
	}
	return ret
}

// Concurrent safe to update rules
func onRuleUpdate(rules []Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()
	newBreakerRules := make(map[string][]Rule)

	for _, rule := range rules {
		if !rule.isApplicable() {
			logger.Warnf("Ignoring invalid breaker rule when loading new rules, %+v.", rule)
			continue
		}

		classification := rule.ResourceName()
		ruleSet, ok := newBreakerRules[classification]
		if !ok {
			ruleSet = make([]Rule, 0, 1)
		}
		ruleSet = append(ruleSet, rule)
		newBreakerRules[classification] = ruleSet
	}

	newBreakers := make(map[string][]CircuitBreaker)
	for k, v := range newBreakerRules {
		cbs := make([]CircuitBreaker, 0, len(v))
		for _, r := range v {
			cbs = append(cbs, r.convert2CircuitBreaker())
		}
		newBreakers[k] = cbs
	}

	ruleMux.Lock()
	breakerRules = newBreakerRules
	breakers = newBreakers
	ruleMux.Unlock()

	logRuleUpdate(newBreakerRules)
	return nil
}

func rulesFrom(rm map[string][]Rule) []Rule {
	rules := make([]Rule, 0)
	if len(rm) == 0 {
		return rules
	}
	for _, rs := range rm {
		if len(rs) == 0 {
			continue
		}
		for _, r := range rs {
			if r != nil {
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func logRuleUpdate(rules map[string][]Rule) {
	sb := strings.Builder{}
	sb.WriteString("[CircuitBreakerRuleManager] succeed to load circuit breakers:\n")
	for _, rule := range rulesFrom(rules) {
		sb.WriteString(rule.String())
		sb.WriteString("\n")
	}
	logger.Info(sb.String())
}
