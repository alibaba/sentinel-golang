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

package flow

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/cluster/server"
	"github.com/alibaba/sentinel-golang/core/config"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// TrafficControllerMap represents the map storage for TrafficShapingController.
type TrafficControllerMap map[string][]*TrafficShapingController

var (
	tcMux = new(sync.RWMutex)
	tcMap = make(TrafficControllerMap)

	currentRules  = make([]*ClusterRule, 0)
	updateRuleMux = new(sync.Mutex)
)

func LoadRules(rules []*ClusterRule) (bool, error) {
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, rules)
	if isEqual {
		logging.Info("[Flow] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}

	err := onRuleUpdate(rules)
	return true, err
}

func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for cluster flow module if user changes the rule.
func GetRules() []ClusterRule {
	rules := getRules()
	ret := make([]ClusterRule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func GetRulesOfResource(res string) []ClusterRule {
	rules := getRulesOfResource(res)
	ret := make([]ClusterRule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func IsValidRule(rule *ClusterRule) error {
	if rule == nil {
		return errors.New("nil Rule")
	}
	if len(rule.Resource) == 0 {
		return errors.New("empty Resource")
	}
	if rule.Threshold < 0 {
		return errors.New("negative Threshold")
	}
	if rule.StatIntervalInMs == 0 {
		return errors.New("zero StatIntervalInMs")
	}
	return nil
}

func onRuleUpdate(rules []*ClusterRule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	resRulesMap := make(map[string][]*ClusterRule, len(rules))
	for _, rule := range rules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("[Cluster Flow onRuleUpdate] Ignoring invalid flow rule", "rule", rule, "reason", err.Error())
			continue
		}
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*ClusterRule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}
	m := make(TrafficControllerMap, len(resRulesMap))
	start := util.CurrentTimeNano()
	for res, rulesOfRes := range resRulesMap {
		m[res] = buildResourceTrafficShapingController(rulesOfRes)
	}

	tcMux.Lock()
	tcMap = m
	currentRules = rules
	tcMux.Unlock()

	logging.Debug("[Cluster Flow onRuleUpdate] Time statistic(ns) for updating flow rule", "timeCost", util.CurrentTimeNano()-start)
	logRuleUpdate(resRulesMap)
	return nil
}

func buildResourceTrafficShapingController(rulesOfRes []*ClusterRule) []*TrafficShapingController {
	newTcsOfRes := make([]*TrafficShapingController, 0, len(rulesOfRes))
	for _, r := range rulesOfRes {
		tc, err := generateTrafficShapingController(r)
		if err != nil {
			logging.Error(err, "[Cluster Flow buildResourceTrafficShapingController] Fail to generate TrafficShapingController", "rule", r)
			continue
		}
		newTcsOfRes = append(newTcsOfRes, tc)
	}
	return newTcsOfRes
}

func logRuleUpdate(m map[string][]*ClusterRule) {
	rules := make([]*ClusterRule, 0, 8)
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		rules = append(rules, rs...)
	}
	if len(rules) == 0 {
		logging.Info("[ClusterFlowRuleManager] Cluster Flow rules were cleared")
	} else {
		logging.Info("[ClusterFlowRuleManager] Cluster Flow rules were loaded", "rules", rules)
	}
}

func generateTrafficShapingController(r *ClusterRule) (*TrafficShapingController, error) {
	if r == nil {
		return nil, errors.New("nil ClusterRule")
	}
	tokenService, err := server.GetTokenService(config.TokenServiceType())
	if err != nil {
		return nil, err
	}
	return &TrafficShapingController{
		rule:         r,
		tokenService: tokenService,
	}, nil
}

func getTrafficControllerListFor(name string) []*TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[name]
}

// getRules returns all the rules。Any changes of rules take effect for flow module
// getRules is an internal interface.
func getRules() []*ClusterRule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return rulesFrom(tcMap)
}

func rulesFrom(m TrafficControllerMap) []*ClusterRule {
	rules := make([]*ClusterRule, 0, 8)
	if len(m) == 0 {
		return rules
	}
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		for _, r := range rs {
			if r != nil && r.BoundRule() != nil {
				rules = append(rules, r.BoundRule())
			}
		}
	}
	return rules
}

// getRulesOfResource returns specific resource's rules。Any changes of rules take effect for flow module
// getRulesOfResource is an internal interface.
func getRulesOfResource(res string) []*ClusterRule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	resTcs, exist := tcMap[res]
	if !exist {
		return nil
	}
	ret := make([]*ClusterRule, 0, len(resTcs))
	for _, tc := range resTcs {
		ret = append(ret, tc.BoundRule())
	}
	return ret
}
