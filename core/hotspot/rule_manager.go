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

package hotspot

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController

// trafficControllerMap represents the map storage for TrafficShapingController.
type trafficControllerMap map[string][]TrafficShapingController

var (
	tcGenFuncMap  = make(map[ControlBehavior]TrafficControllerGenFunc, 4)
	tcMap         = make(trafficControllerMap)
	tcMux         = new(sync.RWMutex)
	currentRules  = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.Mutex)
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[Reject] = func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController {
		var baseTc *baseTrafficShapingController
		if reuseMetric != nil {
			// new BaseTrafficShapingController with reuse statistic metric
			baseTc = newBaseTrafficShapingControllerWithMetric(r, reuseMetric)
		} else {
			baseTc = newBaseTrafficShapingController(r)
		}
		if baseTc == nil {
			return nil
		}
		return &rejectTrafficShapingController{
			baseTrafficShapingController: *baseTc,
			burstCount:                   r.BurstCount,
		}
	}

	tcGenFuncMap[Throttling] = func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController {
		var baseTc *baseTrafficShapingController
		if reuseMetric != nil {
			baseTc = newBaseTrafficShapingControllerWithMetric(r, reuseMetric)
		} else {
			baseTc = newBaseTrafficShapingController(r)
		}
		if baseTc == nil {
			return nil
		}
		return &throttlingTrafficShapingController{
			baseTrafficShapingController: *baseTc,
			maxQueueingTimeMs:            r.MaxQueueingTimeMs,
		}
	}
}

func getTrafficControllersFor(res string) []TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[res]
}

// LoadRules replaces all old hotspot param flow rules with the given rules.
// Return value:
//
//	bool: indicates whether the internal map has been changed;
//	error: indicates whether occurs the error.
func LoadRules(rules []*Rule) (bool, error) {
	resRulesMap := make(map[string][]*Rule, 16)
	for _, rule := range rules {
		resRules, exists := resRulesMap[rule.Resource]
		if !exists {
			resRules = make([]*Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, resRulesMap)
	if isEqual {
		logging.Info("[HotSpot] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}

	err := onRuleUpdate(resRulesMap)
	return true, err
}

// GetRules returns all the hotspot param flow rules based on copy.
// It doesn't take effect for hotspot module if user changes the returned rules.
// GetRules need to compete hotspot module's global lock and the high performance losses of copy,
// reduce or do not call GetRules if possible.
func GetRules() []Rule {
	tcMux.RLock()
	rules := rulesFrom(tcMap)
	tcMux.RUnlock()

	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource returns specific resource's hotspot parameter flow control rules based on copy.
// It doesn't take effect for hotspot module if user changes the returned rules.
// GetRulesOfResource need to compete hotspot module's global lock and the high performance losses of copy,
//
//	reduce or do not call GetRulesOfResource frequently if possible.
func GetRulesOfResource(res string) []Rule {
	tcMux.RLock()
	resTcs := tcMap[res]
	tcMux.RUnlock()

	ret := make([]Rule, 0, len(resTcs))
	for _, tc := range resTcs {
		ret = append(ret, *tc.BoundRule())
	}
	return ret
}

// ClearRules clears all hotspot param flow rules.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// ClearRulesOfResource clears resource level hotspot param flow rules.
func ClearRulesOfResource(res string) error {
	_, err := LoadRulesOfResource(res, nil)
	return err
}

func onRuleUpdate(rawResRulesMap map[string][]*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	// ignore invalid rules
	validResRulesMap := make(map[string][]*Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*Rule, 0, len(rules))
		for _, rule := range rules {
			if err := IsValidRule(rule); err != nil {
				logging.Warn("[HotSpot onRuleUpdate] Ignoring invalid hotspot param flow rule when loading new rules", "rule", rule, "err", err.Error())
				continue
			}
			validResRules = append(validResRules, rule)
		}
		if len(validResRules) > 0 {
			validResRulesMap[res] = validResRules
		}
	}

	start := util.CurrentTimeNano()

	tcMux.RLock()
	tcMapClone := make(trafficControllerMap, len(tcMap))
	for res, tcs := range tcMap {
		resTcClone := make([]TrafficShapingController, 0, len(tcs))
		resTcClone = append(resTcClone, tcs...)
		tcMapClone[res] = resTcClone
	}
	tcMux.RUnlock()

	m := make(trafficControllerMap, len(validResRulesMap))
	for res, rules := range validResRulesMap {
		m[res] = buildResourceTrafficShapingController(res, rules, tcMapClone[res])
	}

	tcMux.Lock()
	tcMap = m
	tcMux.Unlock()

	currentRules = rawResRulesMap

	logging.Debug("[HotSpot onRuleUpdate] Time statistic(ns) for updating hotspot param flow rules", "timeCost", util.CurrentTimeNano()-start)
	logRuleUpdate(validResRulesMap)
	return nil
}

func onResourceRuleUpdate(res string, rawResRules []*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	validResRules := make([]*Rule, 0, len(rawResRules))
	for _, rule := range rawResRules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("[HotSpot onResourceRuleUpdate] Ignoring invalid hotspot param flow rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}

	start := util.CurrentTimeNano()
	oldResTcs := make([]TrafficShapingController, 0, 8)
	tcMux.RLock()
	oldResTcs = append(oldResTcs, tcMap[res]...)
	tcMux.RUnlock()

	newResTcs := buildResourceTrafficShapingController(res, validResRules, oldResTcs)

	tcMux.Lock()
	if len(newResTcs) == 0 {
		delete(tcMap, res)
	} else {
		tcMap[res] = newResTcs
	}
	tcMux.Unlock()

	currentRules[res] = rawResRules

	logging.Debug("[HotSpot onResourceRuleUpdate] Time statistic(ns) for updating hotspot param flow rules", "timeCost", util.CurrentTimeNano()-start)
	logging.Info("[HotSpot] load resource level hotspot param flow rules", "resource", res, "validResRules", validResRules)
	return nil
}

// LoadRulesOfResource loads the given resource's hotspot param flow rules to the rule manager,
// while all previous resource's rules will be replaced. The first returned value indicates whether
// do real load operation, if the rules is the same with previous resource's rules, return false.
func LoadRulesOfResource(res string, rules []*Rule) (bool, error) {
	if len(res) == 0 {
		return false, errors.New("empty resource")
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()

	// clear resource rules
	if len(rules) == 0 {
		// clear resource's currentRules
		delete(currentRules, res)
		// clear tcMap
		tcMux.Lock()
		delete(tcMap, res)
		tcMux.Unlock()
		logging.Info("[HotSpot] clear resource level hotspot param flow rules", "resource", res)
		return true, nil
	}

	// load resource level rules
	isEqual := reflect.DeepEqual(currentRules[res], rules)
	if isEqual {
		logging.Info("[HotSpot] Load resource level hotspot param flow rules is the same with current resource level rules, so ignore load operation.")
		return false, nil
	}

	err := onResourceRuleUpdate(res, rules)
	return true, err
}

func logRuleUpdate(m map[string][]*Rule) {
	rules := make([]*Rule, 0, 8)
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		rules = append(rules, rs...)
	}
	if len(rules) == 0 {
		logging.Info("[HotspotRuleManager] Hotspot param flow rules were cleared")
	} else {
		logging.Info("[HotspotRuleManager] Hotspot param flow rules were loaded", "rules", rules)
	}
}

func rulesFrom(m trafficControllerMap) []*Rule {
	rules := make([]*Rule, 0, 8)
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

func calculateReuseIndexFor(r *Rule, oldResTcs []TrafficShapingController) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old traffic shaping controller slice
	equalIdx = -1
	// the index of statistic reusable rule in old traffic shaping controller slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResTcs {
		oldRule := oldTc.BoundRule()
		if oldRule.Equals(r) {
			// break if there is equivalent rule
			equalIdx = idx
			break
		}
		// find the index of first StatReusable rule
		if !oldRule.IsStatReusable(r) {
			continue
		}
		if reuseStatIdx >= 0 {
			// had find reuse rule.
			continue
		}
		reuseStatIdx = idx
	}
	return equalIdx, reuseStatIdx
}

// buildResourceTrafficShapingController builds TrafficShapingController slice from rules. the resource of rules must be equals to res.
func buildResourceTrafficShapingController(res string, resRules []*Rule, oldResTcs []TrafficShapingController) []TrafficShapingController {
	newTcsOfRes := make([]TrafficShapingController, 0, len(resRules))
	for _, rule := range resRules {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name, expect: %s, actual: %s", res, rule.Resource), "Unmatched resource name in hotspot.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}

		equalIdx, reuseStatIdx := calculateReuseIndexFor(rule, oldResTcs)
		// there is equivalent rule in old traffic shaping controller slice
		if equalIdx >= 0 {
			equalOldTC := oldResTcs[equalIdx]
			newTcsOfRes = append(newTcsOfRes, equalOldTC)
			// remove old tc from old resTcs
			oldResTcs = append(oldResTcs[:equalIdx], oldResTcs[equalIdx+1:]...)
			continue
		}

		// generate new traffic shaping controller
		generator, supported := tcGenFuncMap[rule.ControlBehavior]
		if !supported {
			logging.Warn("[HotSpot buildResourceTrafficShapingController] Ignoring the hotspot param flow rule due to unsupported control behavior", "rule", rule)
			continue
		}
		var tc TrafficShapingController
		if reuseStatIdx >= 0 {
			// generate new traffic shaping controller with reusable statistic metric.
			tc = generator(rule, oldResTcs[reuseStatIdx].BoundMetric())
			// remove the reused traffic shaping controller old res tcs
			oldResTcs = append(oldResTcs[:reuseStatIdx], oldResTcs[reuseStatIdx+1:]...)
		} else {
			tc = generator(rule, nil)
		}
		if tc == nil {
			logging.Debug("[HotSpot buildResourceTrafficShapingController] Ignoring the hotspot param flow rule due to bad generated traffic controller", "rule", rule)
			continue
		}

		newTcsOfRes = append(newTcsOfRes, tc)
	}
	return newTcsOfRes
}

func IsValidRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil hotspot Rule")
	}
	if len(rule.Resource) == 0 {
		return errors.New("empty resource name")
	}
	if rule.Threshold < 0 {
		return errors.New("negative threshold")
	}
	if rule.MetricType < 0 {
		return errors.New("invalid metric type")
	}
	if rule.ControlBehavior < 0 {
		return errors.New("invalid control strategy")
	}
	if rule.MetricType == QPS && rule.DurationInSec <= 0 {
		return errors.New("invalid duration")
	}
	if rule.ParamIndex > 0 && rule.ParamKey != "" {
		return errors.New("invalid param index and param key are mutually exclusive")
	}
	return checkControlBehaviorField(rule)
}

func checkControlBehaviorField(rule *Rule) error {
	switch rule.ControlBehavior {
	case Reject:
		if rule.BurstCount < 0 {
			return errors.New("invalid BurstCount")
		}
		return nil
	case Throttling:
		if rule.MaxQueueingTimeMs < 0 {
			return errors.New("invalid MaxQueueingTimeMs")
		}
		return nil
	default:
	}
	return nil
}

// SetTrafficShapingGenerator sets the traffic controller generator for the given control behavior.
// Note that modifying the generator of default control behaviors is not allowed.
func SetTrafficShapingGenerator(cb ControlBehavior, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if cb >= Reject && cb <= Throttling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[cb] = generator
	return nil
}

func RemoveTrafficShapingGenerator(cb ControlBehavior) error {
	if cb >= Reject && cb <= Throttling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, cb)
	return nil
}
