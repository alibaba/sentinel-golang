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
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/stat"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(*Rule, *standaloneStatistic) (*TrafficShapingController, error)

type trafficControllerGenKey struct {
	tokenCalculateStrategy TokenCalculateStrategy
	controlBehavior        ControlBehavior
}

// TrafficControllerMap represents the map storage for TrafficShapingController.
type TrafficControllerMap map[string][]*TrafficShapingController

var (
	tcGenFuncMap = make(map[trafficControllerGenKey]TrafficControllerGenFunc, 6)
	tcMap        = make(TrafficControllerMap)
	tcMux        = new(sync.RWMutex)
	nopStat      = &standaloneStatistic{
		reuseResourceStat: false,
		readOnlyMetric:    base.NopReadStat(),
		writeOnlyMetric:   base.NopWriteStat(),
	}
	currentRules  = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.Mutex)
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: Direct,
		controlBehavior:        Reject,
	}] = func(rule *Rule, boundStat *standaloneStatistic) (*TrafficShapingController, error) {
		if boundStat == nil {
			var err error
			boundStat, err = generateStatFor(rule)
			if err != nil {
				return nil, err
			}
		}
		tsc, err := NewTrafficShapingController(rule, boundStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewDirectTrafficShapingCalculator(tsc, rule.Threshold)
		tsc.flowChecker = NewRejectTrafficShapingChecker(tsc, rule)
		return tsc, nil
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: Direct,
		controlBehavior:        Throttling,
	}] = func(rule *Rule, _ *standaloneStatistic) (*TrafficShapingController, error) {
		// Direct token calculate strategy and throttling control behavior don't use stat, so we just give a nop stat.
		tsc, err := NewTrafficShapingController(rule, nopStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewDirectTrafficShapingCalculator(tsc, rule.Threshold)
		tsc.flowChecker = NewThrottlingChecker(tsc, rule.MaxQueueingTimeMs, rule.StatIntervalInMs)
		return tsc, nil
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: WarmUp,
		controlBehavior:        Reject,
	}] = func(rule *Rule, boundStat *standaloneStatistic) (*TrafficShapingController, error) {
		if boundStat == nil {
			var err error
			boundStat, err = generateStatFor(rule)
			if err != nil {
				return nil, err
			}
		}
		tsc, err := NewTrafficShapingController(rule, boundStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewWarmUpTrafficShapingCalculator(tsc, rule)
		tsc.flowChecker = NewRejectTrafficShapingChecker(tsc, rule)
		return tsc, nil
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: WarmUp,
		controlBehavior:        Throttling,
	}] = func(rule *Rule, boundStat *standaloneStatistic) (*TrafficShapingController, error) {
		if boundStat == nil {
			var err error
			boundStat, err = generateStatFor(rule)
			if err != nil {
				return nil, err
			}
		}
		tsc, err := NewTrafficShapingController(rule, boundStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewWarmUpTrafficShapingCalculator(tsc, rule)
		tsc.flowChecker = NewThrottlingChecker(tsc, rule.MaxQueueingTimeMs, rule.StatIntervalInMs)
		return tsc, nil
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: MemoryAdaptive,
		controlBehavior:        Reject,
	}] = func(rule *Rule, boundStat *standaloneStatistic) (*TrafficShapingController, error) {
		if boundStat == nil {
			var err error
			boundStat, err = generateStatFor(rule)
			if err != nil {
				return nil, err
			}
		}
		tsc, err := NewTrafficShapingController(rule, boundStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewMemoryAdaptiveTrafficShapingCalculator(tsc, rule)
		tsc.flowChecker = NewRejectTrafficShapingChecker(tsc, rule)
		return tsc, nil
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: MemoryAdaptive,
		controlBehavior:        Throttling,
	}] = func(rule *Rule, _ *standaloneStatistic) (*TrafficShapingController, error) {
		// MemoryAdaptive token calculate strategy and throttling control behavior don't use stat, so we just give a nop stat.
		tsc, err := NewTrafficShapingController(rule, nopStat)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewMemoryAdaptiveTrafficShapingCalculator(tsc, rule)
		tsc.flowChecker = NewThrottlingChecker(tsc, rule.MaxQueueingTimeMs, rule.StatIntervalInMs)
		return tsc, nil
	}
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
		logging.Info("[FlowRuleManager] Flow rules were cleared")
	} else {
		logging.Info("[FlowRuleManager] Flow rules were loaded", "rules", rules)
	}
}

func onRuleUpdate(rawResRulesMap map[string][]*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	// ignore invalid rules
	validResRulesMap := make(map[string][]*Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*Rule, 0, len(rules))
		for _, rule := range rules {
			if err := IsValidRule(rule); err != nil {
				logging.Warn("[Flow onRuleUpdate] Ignoring invalid flow rule", "rule", rule, "reason", err.Error())
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
	tcMapClone := make(TrafficControllerMap, len(validResRulesMap))
	for res, tcs := range tcMap {
		resTcClone := make([]*TrafficShapingController, 0, len(tcs))
		resTcClone = append(resTcClone, tcs...)
		tcMapClone[res] = resTcClone
	}
	tcMux.RUnlock()

	m := make(TrafficControllerMap, len(validResRulesMap))
	for res, rulesOfRes := range validResRulesMap {
		newTcsOfRes := buildResourceTrafficShapingController(res, rulesOfRes, tcMapClone[res])
		if len(newTcsOfRes) > 0 {
			m[res] = newTcsOfRes
		}
	}

	tcMux.Lock()
	tcMap = m
	tcMux.Unlock()
	currentRules = rawResRulesMap

	logging.Debug("[Flow onRuleUpdate] Time statistic(ns) for updating flow rule", "timeCost", util.CurrentTimeNano()-start)
	logRuleUpdate(validResRulesMap)
	return nil
}

// LoadRules loads the given flow rules to the rule manager, while all previous rules will be replaced.
// the first returned value indicates whether do real load operation, if the rules is the same with previous rules, return false
func LoadRules(rules []*Rule) (bool, error) {
	resRulesMap := make(map[string][]*Rule, 16)
	for _, rule := range rules {
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, resRulesMap)
	if isEqual {
		logging.Info("[Flow] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(resRulesMap)
	return true, err
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
			logging.Warn("[Flow onResourceRuleUpdate] Ignoring invalid flow rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}

	start := util.CurrentTimeNano()
	oldResTcs := make([]*TrafficShapingController, 0)
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
	logging.Debug("[Flow onResourceRuleUpdate] Time statistic(ns) for updating flow rule", "timeCost", util.CurrentTimeNano()-start)
	logging.Info("[Flow] load resource level rules", "resource", res, "validResRules", validResRules)
	return nil
}

// LoadRulesOfResource loads the given resource's flow rules to the rule manager, while all previous resource's rules will be replaced.
// the first returned value indicates whether do real load operation, if the rules is the same with previous resource's rules, return false
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
		logging.Info("[Flow] clear resource level rules", "resource", res)
		return true, nil
	}
	// load resource level rules
	isEqual := reflect.DeepEqual(currentRules[res], rules)
	if isEqual {
		logging.Info("[Flow] Load resource level rules is the same with current resource level rules, so ignore load operation.")
		return false, nil
	}

	err := onResourceRuleUpdate(res, rules)
	return true, err
}

// getRules returns all the rules。Any changes of rules take effect for flow module
// getRules is an internal interface.
func getRules() []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return rulesFrom(tcMap)
}

// getRulesOfResource returns specific resource's rules。Any changes of rules take effect for flow module
// getRulesOfResource is an internal interface.
func getRulesOfResource(res string) []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	resTcs, exist := tcMap[res]
	if !exist {
		return nil
	}
	ret := make([]*Rule, 0, len(resTcs))
	for _, tc := range resTcs {
		ret = append(ret, tc.BoundRule())
	}
	return ret
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for flow module if user changes the rule.
func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource returns specific resource's rules based on copy.
// It doesn't take effect for flow module if user changes the rule.
func GetRulesOfResource(res string) []Rule {
	rules := getRulesOfResource(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// ClearRules clears all the rules in flow module.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// ClearRulesOfResource clears resource level rules in flow module.
func ClearRulesOfResource(res string) error {
	_, err := LoadRulesOfResource(res, nil)
	return err
}

func rulesFrom(m TrafficControllerMap) []*Rule {
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

func generateStatFor(rule *Rule) (*standaloneStatistic, error) {
	if !rule.needStatistic() {
		return nopStat, nil
	}

	intervalInMs := rule.StatIntervalInMs

	var retStat standaloneStatistic

	var resNode *stat.ResourceNode
	if rule.RelationStrategy == AssociatedResource {
		// use associated statistic
		resNode = stat.GetOrCreateResourceNode(rule.RefResource, base.ResTypeCommon)
	} else {
		resNode = stat.GetOrCreateResourceNode(rule.Resource, base.ResTypeCommon)
	}
	if intervalInMs == 0 || intervalInMs == config.MetricStatisticIntervalMs() {
		// default case, use the resource's default statistic
		readStat := resNode.DefaultMetric()
		retStat.reuseResourceStat = true
		retStat.readOnlyMetric = readStat
		retStat.writeOnlyMetric = nil
		return &retStat, nil
	}

	sampleCount := uint32(0)
	//calculate the sample count
	if intervalInMs > config.GlobalStatisticIntervalMsTotal() {
		sampleCount = 1
	} else if intervalInMs < config.GlobalStatisticBucketLengthInMs() {
		sampleCount = 1
	} else {
		if intervalInMs%config.GlobalStatisticBucketLengthInMs() == 0 {
			sampleCount = intervalInMs / config.GlobalStatisticBucketLengthInMs()
		} else {
			sampleCount = 1
		}
	}
	err := base.CheckValidityForReuseStatistic(sampleCount, intervalInMs, config.GlobalStatisticSampleCountTotal(), config.GlobalStatisticIntervalMsTotal())
	if err == nil {
		// global statistic reusable
		readStat, e := resNode.GenerateReadStat(sampleCount, intervalInMs)
		if e != nil {
			return nil, e
		}
		retStat.reuseResourceStat = true
		retStat.readOnlyMetric = readStat
		retStat.writeOnlyMetric = nil
		return &retStat, nil
	} else if err == base.GlobalStatisticNonReusableError {
		logging.Info("[FlowRuleManager] Flow rule couldn't reuse global statistic and will generate independent statistic", "rule", rule)
		retStat.reuseResourceStat = false
		realLeapArray := sbase.NewBucketLeapArray(sampleCount, intervalInMs)
		metricStat, e := sbase.NewSlidingWindowMetric(sampleCount, intervalInMs, realLeapArray)
		if e != nil {
			return nil, errors.Errorf("fail to generate statistic for warm up rule: %+v, err: %+v", rule, e)
		}
		retStat.readOnlyMetric = metricStat
		retStat.writeOnlyMetric = realLeapArray
		return &retStat, nil
	}
	return nil, errors.Wrapf(err, "fail to new standalone statistic because of invalid StatIntervalInMs in flow.Rule, StatIntervalInMs: %d", intervalInMs)
}

// SetTrafficShapingGenerator sets the traffic controller generator for the given TokenCalculateStrategy and ControlBehavior.
// Note that modifying the generator of default control strategy is not allowed.
func SetTrafficShapingGenerator(tokenCalculateStrategy TokenCalculateStrategy, controlBehavior ControlBehavior, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}

	if tokenCalculateStrategy >= Direct && tokenCalculateStrategy <= WarmUp {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	if controlBehavior >= Reject && controlBehavior <= Throttling {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: tokenCalculateStrategy,
		controlBehavior:        controlBehavior,
	}] = generator
	return nil
}

func RemoveTrafficShapingGenerator(tokenCalculateStrategy TokenCalculateStrategy, controlBehavior ControlBehavior) error {
	if tokenCalculateStrategy >= Direct && tokenCalculateStrategy <= WarmUp {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	if controlBehavior >= Reject && controlBehavior <= Throttling {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, trafficControllerGenKey{
		tokenCalculateStrategy: tokenCalculateStrategy,
		controlBehavior:        controlBehavior,
	})
	return nil
}

func getTrafficControllerListFor(name string) []*TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[name]
}

func calculateReuseIndexFor(r *Rule, oldResTcs []*TrafficShapingController) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old traffic shaping controller slice
	equalIdx = -1
	// the index of statistic reusable rule in old traffic shaping controller slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResTcs {
		oldRule := oldTc.BoundRule()
		if oldRule.isEqualsTo(r) {
			// break if there is equivalent rule
			equalIdx = idx
			break
		}
		// search the index of first stat reusable rule
		if !oldRule.isStatReusable(r) {
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

// buildResourceTrafficShapingController builds TrafficShapingController slice from rules. the resource of rules must be equals to res
func buildResourceTrafficShapingController(res string, rulesOfRes []*Rule, oldResTcs []*TrafficShapingController) []*TrafficShapingController {
	newTcsOfRes := make([]*TrafficShapingController, 0, len(rulesOfRes))
	for _, rule := range rulesOfRes {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name expect: %s, actual: %s", res, rule.Resource), "Unmatched resource name in flow.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		equalIdx, reuseStatIdx := calculateReuseIndexFor(rule, oldResTcs)

		// First check equals scenario
		if equalIdx >= 0 {
			// reuse the old tc
			equalOldTc := oldResTcs[equalIdx]
			newTcsOfRes = append(newTcsOfRes, equalOldTc)
			// remove old tc from oldResTcs
			oldResTcs = append(oldResTcs[:equalIdx], oldResTcs[equalIdx+1:]...)
			continue
		}

		generator, supported := tcGenFuncMap[trafficControllerGenKey{
			tokenCalculateStrategy: rule.TokenCalculateStrategy,
			controlBehavior:        rule.ControlBehavior,
		}]
		if !supported || generator == nil {
			logging.Error(errors.New("unsupported flow control strategy"), "Ignoring the rule due to unsupported control behavior in flow.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		var tc *TrafficShapingController
		var e error
		if reuseStatIdx >= 0 {
			tc, e = generator(rule, &(oldResTcs[reuseStatIdx].boundStat))
		} else {
			tc, e = generator(rule, nil)
		}

		if tc == nil || e != nil {
			logging.Error(errors.New("bad generated traffic controller"), "Ignoring the rule due to bad generated traffic controller in flow.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		if reuseStatIdx >= 0 {
			// remove old tc from oldResTcs
			oldResTcs = append(oldResTcs[:reuseStatIdx], oldResTcs[reuseStatIdx+1:]...)
		}
		newTcsOfRes = append(newTcsOfRes, tc)
	}
	return newTcsOfRes
}

// IsValidRule checks whether the given Rule is valid.
func IsValidRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil Rule")
	}
	if rule.Resource == "" {
		return errors.New("empty Resource")
	}
	if rule.Threshold < 0 {
		return errors.New("negative Threshold")
	}
	if int32(rule.TokenCalculateStrategy) < 0 {
		return errors.New("negative TokenCalculateStrategy")
	}
	if int32(rule.ControlBehavior) < 0 {
		return errors.New("negative ControlBehavior")
	}
	if !(rule.RelationStrategy >= CurrentResource && rule.RelationStrategy <= AssociatedResource) {
		return errors.New("invalid RelationStrategy")
	}
	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return errors.New("RefResource must be non empty when RelationStrategy is AssociatedResource")
	}
	if rule.TokenCalculateStrategy == WarmUp {
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("WarmUpPeriodSec must be great than 0")
		}
		if rule.WarmUpColdFactor == 1 {
			return errors.New("WarmUpColdFactor must be great than 1")
		}
	}
	if rule.StatIntervalInMs > 10*60*1000 {
		logging.Info("StatIntervalInMs is great than 10 minutes, less than 10 minutes is recommended.")
	}
	if rule.TokenCalculateStrategy == MemoryAdaptive {
		if rule.LowMemUsageThreshold <= 0 {
			return errors.New("rule.LowMemUsageThreshold <= 0")
		}
		if rule.HighMemUsageThreshold <= 0 {
			return errors.New("rule.HighMemUsageThreshold <= 0")
		}
		if rule.HighMemUsageThreshold >= rule.LowMemUsageThreshold {
			return errors.New("rule.HighMemUsageThreshold >= rule.LowMemUsageThreshold")
		}

		if rule.MemLowWaterMarkBytes <= 0 {
			return errors.New("rule.MemLowWaterMarkBytes <= 0")
		}
		if rule.MemHighWaterMarkBytes <= 0 {
			return errors.New("rule.MemHighWaterMarkBytes <= 0")
		}
		if rule.MemHighWaterMarkBytes > int64(system_metric.TotalMemorySize) {
			return errors.New("rule.MemHighWaterMarkBytes should not be greater than current system's total memory size")
		}
		if rule.MemLowWaterMarkBytes >= rule.MemHighWaterMarkBytes {
			// can not be equal to defeat from zero overflow
			return errors.New("rule.MemLowWaterMarkBytes >= rule.MemHighWaterMarkBytes")
		}
	}

	return nil
}
