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

	currentRules = make(map[string]*Rule)
	updateMux    = new(sync.RWMutex)
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
	isEqual := reflect.DeepEqual(currentRules, rulesMap)
	if isEqual {
		logging.Info("[Outlier] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(rulesMap)
	return true, err
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
		err := IsValidRule(rule)
		err = circuitbreaker.IsValidRule(circuitRule)
		if err != nil {
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
