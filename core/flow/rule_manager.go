package flow

import (
	"fmt"
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/stat"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
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
	tcGenFuncMap = make(map[trafficControllerGenKey]TrafficControllerGenFunc)
	tcMap        = make(TrafficControllerMap)
	tcMux        = new(sync.RWMutex)
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
		tsc.flowChecker = NewThrottlingChecker(tsc, rule.MaxQueueingTimeMs)
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
		tsc.flowChecker = NewThrottlingChecker(tsc, rule.MaxQueueingTimeMs)
		return tsc, nil
	}
}

func logRuleUpdate(m TrafficControllerMap) {
	rs := rulesFrom(m)
	if len(rs) == 0 {
		logging.Info("[FlowRuleManager] Flow rules were cleared")
	} else {
		logging.Info("[FlowRuleManager] Flow rules were loaded", "rules", rs)
	}
}

func onRuleUpdate(rules []*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	resRulesMap := make(map[string][]*Rule)
	for _, rule := range rules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("ignoring invalid flow rule", "rule", rule, "reason", err)
			continue
		}
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}
	m := make(TrafficControllerMap, len(resRulesMap))
	start := util.CurrentTimeNano()
	tcMux.Lock()
	defer func() {
		tcMux.Unlock()
		if r := recover(); r != nil {
			return
		}
		logging.Debug("time statistic(ns) for updating flow rule", "timeCost", util.CurrentTimeNano()-start)
		logRuleUpdate(m)
	}()
	for res, rulesOfRes := range resRulesMap {
		m[res] = buildRulesOfRes(res, rulesOfRes)
	}
	tcMap = m
	return nil
}

// LoadRules loads the given flow rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*Rule) (bool, error) {
	// TODO: rethink the design
	err := onRuleUpdate(rules)
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

func rulesFrom(m TrafficControllerMap) []*Rule {
	rules := make([]*Rule, 0)
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
		logging.Info("flow rule couldn't reuse global statistic and will generate independent statistic", "rule", rule)
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

func calculateReuseIndexFor(r *Rule, oldResCbs []*TrafficShapingController) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old circuit breaker slice
	equalIdx = -1
	// the index of statistic reusable rule in old circuit breaker slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResCbs {
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

// buildRulesOfRes builds TrafficShapingController slice from rules. the resource of rules must be equals to res
func buildRulesOfRes(res string, rulesOfRes []*Rule) []*TrafficShapingController {
	newTcsOfRes := make([]*TrafficShapingController, 0, len(rulesOfRes))
	emptyTcs := make([]*TrafficShapingController, 0, 0)
	for _, rule := range rulesOfRes {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name, expect: %s, actual: %s", res, rule.Resource), "FlowManager: unmatched resource name ", "rule", rule)
			continue
		}
		oldResTcs, exist := tcMap[res]
		if !exist {
			oldResTcs = emptyTcs
		}

		equalIdx, reuseStatIdx := calculateReuseIndexFor(rule, oldResTcs)

		// First check equals scenario
		if equalIdx >= 0 {
			// reuse the old cb
			equalOldTc := oldResTcs[equalIdx]
			newTcsOfRes = append(newTcsOfRes, equalOldTc)
			// remove old cb from oldResCbs
			tcMap[res] = append(oldResTcs[:equalIdx], oldResTcs[equalIdx+1:]...)
			continue
		}

		generator, supported := tcGenFuncMap[trafficControllerGenKey{
			tokenCalculateStrategy: rule.TokenCalculateStrategy,
			controlBehavior:        rule.ControlBehavior,
		}]
		if !supported || generator == nil {
			logging.Error(errors.New("unsupported flow control strategy"), "ignoring the rule due to unsupported control behavior", "rule", rule)
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
			logging.Error(errors.New("bad generated traffic controller"), "ignoring the rule due to bad generated traffic controller.", "rule", rule)
			continue
		}
		if reuseStatIdx >= 0 {
			// remove old cb from oldResCbs
			tcMap[res] = append(oldResTcs[:reuseStatIdx], oldResTcs[reuseStatIdx+1:]...)
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
		return errors.New("empty resource name")
	}
	if rule.Threshold < 0 {
		return errors.New("negative threshold")
	}
	if int32(rule.TokenCalculateStrategy) < 0 {
		return errors.New("invalid token calculate strategy")
	}
	if int32(rule.ControlBehavior) < 0 {
		return errors.New("invalid control behavior")
	}
	if !(rule.RelationStrategy >= CurrentResource && rule.RelationStrategy <= AssociatedResource) {
		return errors.New("invalid relation strategy")
	}
	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return errors.New("Bad flow rule: invalid relation strategy")
	}
	if rule.TokenCalculateStrategy == WarmUp {
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("invalid WarmUpPeriodSec")
		}
		if rule.WarmUpColdFactor == 1 {
			return errors.New("WarmUpColdFactor must be great than 1")
		}
	}
	if rule.ControlBehavior == Throttling && rule.MaxQueueingTimeMs == 0 {
		return errors.New("invalid MaxQueueingTimeMs")
	}
	if rule.StatIntervalInMs > config.GlobalStatisticIntervalMsTotal()*60 {
		return errors.New("StatIntervalInMs must be less than 10 minutes")
	}
	return nil
}
