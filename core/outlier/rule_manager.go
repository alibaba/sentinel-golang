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

package outlier

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

var (
	// resource name ---> outlier ejection rule
	outlierRules = make(map[string]*Rule)
	// resource name ---> circuitbreaker rule
	breakerRules = make(map[string]*circuitbreaker.Rule)
	// resource name ---> address ---> circuitbreaker
	nodeBreakers = make(map[string]map[string]circuitbreaker.CircuitBreaker)
	// resource name ---> outlier ejection rule
	currentRules  = make(map[string]*Rule)
	updateMux     = new(sync.RWMutex)
	updateRuleMux = new(sync.Mutex)
)

func getNodeBreakersOfResource(resource string) map[string]circuitbreaker.CircuitBreaker {
	updateMux.RLock()
	nodes := nodeBreakers[resource]
	updateMux.RUnlock()
	ret := make(map[string]circuitbreaker.CircuitBreaker, len(nodes))
	for address, breaker := range nodes {
		ret[address] = breaker
	}
	return ret
}

func deleteNodeBreakerOfResource(resource string, address string) {
	updateMux.Lock()
	defer updateMux.Unlock()
	if _, ok := nodeBreakers[resource]; ok {
		delete(nodeBreakers[resource], address)
		logging.Info("[Outlier] delete node breaker", "resourceName", resource, "address", address)
	}
}

func addNodeBreakerOfResource(resource string, address string) {
	newBreakers := circuitbreaker.BuildResourceCircuitBreaker(resource,
		[]*circuitbreaker.Rule{getBreakerRuleOfResource(resource)}, []circuitbreaker.CircuitBreaker{})
	if len(newBreakers) > 0 {
		updateMux.Lock()
		if nodeBreakers[resource] == nil {
			nodeBreakers[resource] = make(map[string]circuitbreaker.CircuitBreaker)
		}
		nodeBreakers[resource][address] = newBreakers[0]
		updateMux.Unlock()
		logging.Info("[Outlier] add node breaker", "resourceName", resource, "address", address)
	}
}

func getOutlierRuleOfResource(resource string) *Rule {
	updateMux.RLock()
	rule := outlierRules[resource]
	updateMux.RUnlock()
	return rule
}

func getBreakerRuleOfResource(resource string) *circuitbreaker.Rule {
	updateMux.RLock()
	rule := breakerRules[resource]
	updateMux.RUnlock()
	return rule
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for outlier ejection module if user changes the rule.
// GetRules need to compete outlier ejection module's global lock and the high performance losses of copy,
//
//	reduce or do not call GetRules if possible
func GetRules() []Rule {
	updateMux.RLock()
	rules := rulesFrom(outlierRules)
	updateMux.RUnlock()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func rulesFrom(rm map[string]*Rule) []*Rule {
	rules := make([]*Rule, 0, 8)
	if len(rm) == 0 {
		return rules
	}
	for _, r := range rm {
		if r != nil {
			rules = append(rules, r)
		}
	}
	return rules
}

// ClearRules clear all the previous rules.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// LoadRules replaces old outlier ejection rules with the given rules.
//
// return value:
//
// bool: was designed to indicate whether the internal map has been changed
// error: was designed to indicate whether occurs the error.
func LoadRules(rules []*Rule) (bool, error) {
	rulesMap := make(map[string]*Rule, 16)
	for _, rule := range rules {
		rulesMap[rule.Resource] = rule
	}
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, rulesMap)
	if isEqual {
		logging.Info("[Outlier] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(rulesMap)
	return true, err
}

// LoadRuleOfResource loads the given resource's outlier ejection rule to the rule manager, while previous resource's rule will be replaced.
// the first returned value indicates whether do real load operation, if the rule is the same with previous resource's rule, return false
func LoadRuleOfResource(res string, rule *Rule) (bool, error) {
	if len(res) == 0 {
		return false, errors.New("empty resource")
	}
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	// clear resource rule
	if rule == nil {
		delete(currentRules, res)
		updateMux.Lock()
		delete(nodeBreakers, res)
		delete(breakerRules, res)
		delete(outlierRules, res)
		updateMux.Unlock()
		logging.Info("[Outlier] clear resource level rule", "resource", res)
		return true, nil
	}
	// load resource level rule
	isEqual := reflect.DeepEqual(currentRules[res], rule)
	if isEqual {
		logging.Info("[Outlier] Load resource level rule is the same with current resource level rule, so ignore load operation.")
		return false, nil
	}
	err := onResourceRuleUpdate(res, rule)
	return true, err
}

func onResourceRuleUpdate(res string, rule *Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	circuitRule := rule.Rule
	if err = IsValidRule(rule); err != nil {
		logging.Warn("[Outlier onResourceRuleUpdate] Ignoring invalid outlier ejection rule", "rule", rule, "err", err.Error())
		return
	}
	if err = circuitbreaker.IsValidRule(circuitRule); err != nil {
		logging.Warn("[Outlier onRuleUpdate] Ignoring invalid rule when loading new rules", "rule", rule, "err", err.Error())
		return
	}

	start := util.CurrentTimeNano()
	breakers := getNodeBreakersOfResource(res)
	newBreakers := make(map[string]circuitbreaker.CircuitBreaker)
	for address, breaker := range breakers {
		newCbsOfRes := circuitbreaker.BuildResourceCircuitBreaker(res,
			[]*circuitbreaker.Rule{circuitRule}, []circuitbreaker.CircuitBreaker{breaker})
		if len(newCbsOfRes) > 0 {
			newBreakers[address] = newCbsOfRes[0]
		}
	}

	updateMux.Lock()
	outlierRules[res] = rule
	breakerRules[res] = circuitRule
	nodeBreakers[res] = newBreakers
	updateMux.Unlock()
	currentRules[res] = rule

	logging.Debug("[Outlier onResourceRuleUpdate] Time statistics(ns) for updating outlier ejection rule", "timeCost", util.CurrentTimeNano()-start)
	logging.Info("[Outlier] load resource level rule", "resource", res, "rule", rule)
	return nil
}

// onRuleUpdate is concurrent safe to update outlier ejection rules
func onRuleUpdate(rulesMap map[string]*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	// ignore invalid outlier ejection rule
	validCircuitRulesMap := make(map[string]*circuitbreaker.Rule, len(rulesMap))
	validRulesMap := make(map[string]*Rule, len(rulesMap))
	for resource, rule := range rulesMap {
		circuitRule := rule.Rule
		if err = IsValidRule(rule); err != nil {
			logging.Warn("[Outlier onRuleUpdate] Ignoring invalid rule when loading new rules", "rule", rule, "err", err.Error())
			continue
		}
		if err = circuitbreaker.IsValidRule(circuitRule); err != nil {
			logging.Warn("[Outlier onRuleUpdate] Ignoring invalid rule when loading new rules", "rule", rule, "err", err.Error())
			continue
		}
		validCircuitRulesMap[resource] = circuitRule
		validRulesMap[resource] = rule
	}

	currentRules = rulesMap
	updateMux.Lock()
	breakerRules = validCircuitRulesMap
	outlierRules = validRulesMap
	updateMux.Unlock()

	updateAllBreakers()
	LogRuleUpdate(outlierRules)
	return nil
}

// ClearRuleOfResource clears resource level rule in outlier ejection module.
func ClearRuleOfResource(res string) error {
	_, err := LoadRuleOfResource(res, nil)
	return err
}

func IsValidRule(r *Rule) error {
	if r == nil {
		return errors.New("nil Rule")
	}
	if len(r.Resource) == 0 {
		return errors.New("empty resource name")
	}
	if r.MaxEjectionPercent < 0.0 || r.MaxEjectionPercent > 1.0 {
		return errors.New("invalid MaxEjectionPercent")
	}
	return nil
}

func updateAllBreakers() {
	start := util.CurrentTimeNano()
	updateMux.RLock()
	breakersClone := make(map[string]map[string]circuitbreaker.CircuitBreaker, len(nodeBreakers))
	for resource, breakers := range nodeBreakers {
		breakersClone[resource] = make(map[string]circuitbreaker.CircuitBreaker)
		for address, breaker := range breakers {
			breakersClone[resource][address] = breaker
		}
	}
	updateMux.RUnlock()

	newBreakers := make(map[string]map[string]circuitbreaker.CircuitBreaker, len(breakerRules))
	for resource, rule := range breakerRules {
		newBreakers[resource] = make(map[string]circuitbreaker.CircuitBreaker)
		for address, breaker := range breakersClone[resource] {
			newCbsOfRes := circuitbreaker.BuildResourceCircuitBreaker(resource,
				[]*circuitbreaker.Rule{rule}, []circuitbreaker.CircuitBreaker{breaker})
			if len(newCbsOfRes) > 0 {
				newBreakers[resource][address] = newCbsOfRes[0]
			}
		}
	}

	updateMux.Lock()
	nodeBreakers = newBreakers
	updateMux.Unlock()

	logging.Debug("[Outlier onRuleUpdate] Time statistics(ns) for updating all circuit breakers", "timeCost", util.CurrentTimeNano()-start)
}

func LogRuleUpdate(rules map[string]*Rule) {
	if len(rules) == 0 {
		logging.Info("[OutlierRuleManager] Outlier ejection rules were cleared")
	} else {
		logging.Info("[OutlierRuleManager] Outlier ejection rules were loaded", "rules", rules)
	}
}
