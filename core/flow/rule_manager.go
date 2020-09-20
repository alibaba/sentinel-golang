package flow

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(*Rule) *TrafficShapingController

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
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewDirectTrafficShapingCalculator(rule.Count), NewDefaultTrafficShapingChecker(rule), rule)
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: Direct,
		controlBehavior:        Throttling,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewDirectTrafficShapingCalculator(rule.Count), NewThrottlingChecker(rule.MaxQueueingTimeMs), rule)
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: WarmUp,
		controlBehavior:        Reject,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewWarmUpTrafficShapingCalculator(rule), NewDefaultTrafficShapingChecker(rule), rule)
	}
	tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: WarmUp,
		controlBehavior:        Throttling,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewWarmUpTrafficShapingCalculator(rule), NewThrottlingChecker(rule.MaxQueueingTimeMs), rule)
	}
}

func logRuleUpdate(m TrafficControllerMap) {
	bs, err := json.Marshal(rulesFrom(m))
	if err != nil {
		if len(m) == 0 {
			logging.Info("[FlowRuleManager] Flow rules were cleared")
		} else {
			logging.Info("[FlowRuleManager] Flow rules were loaded")
		}
	} else {
		if len(m) == 0 {
			logging.Info("[FlowRuleManager] Flow rules were cleared")
		} else {
			logging.Info("[FlowRuleManager] Flow rules were loaded", "rules", string(bs))
		}
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

	m := buildRulesMap(rules)

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

	tcMap = m
	return nil
}

// LoadRules loads the given flow rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*Rule) (bool, error) {
	// TODO: rethink the design
	err := onRuleUpdate(rules)
	return true, err
}

// LoadRulesOfResource replaces all the rules of resource level
// First returned value indicates whether replaces
// Second returned value indicates the errors when loads rules
func LoadRulesOfResource(res string, rulesOfRes []*Rule) (bool, error) {
	var loadErr error
	validResRules := make([]*Rule, 0, len(rulesOfRes))
	for _, r := range rulesOfRes {
		if err := IsValidRule(r); err != nil {
			loadErr = multierror.Append(loadErr, err)
			logging.Error(err, "invalid rule.", "rule", r)
			continue
		}
		if r.Resource != res {
			loadErr = multierror.Append(loadErr, errors.Errorf("resource name unmatched, expect: %s, actual: %s", res, r.Resource))
			logging.Error(errors.Errorf("resource name unmatched, expect: %s, actual: %s", res, r.Resource), "unmatched rule.", "rule", r)
			continue
		}
		validResRules = append(validResRules, r)
	}

	resTcs := buildRulesOfRes(res, validResRules)
	loadedRulesMap := make(TrafficControllerMap)
	loadedRulesMap[res] = resTcs
	start := util.CurrentTimeNano()
	tcMux.Lock()
	defer func() {
		tcMux.Unlock()
		if r := recover(); r != nil {
			return
		}
		logging.Debug("time statistic(ns) for loading resource's flow rules", "timeCost", util.CurrentTimeNano()-start)
		logRuleUpdate(loadedRulesMap)
	}()
	tcMap[res] = resTcs
	return true, loadErr

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
		ret = append(ret, tc.Rule())
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
			if r != nil && r.Rule() != nil {
				rules = append(rules, r.Rule())
			}
		}
	}
	return rules
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

// NotThreadSafe (should be guarded by the lock)
func buildRulesMap(rules []*Rule) TrafficControllerMap {
	m := make(TrafficControllerMap)
	if len(rules) == 0 {
		return m
	}

	resRulesMap := make(map[string][]*Rule)
	for _, rule := range rules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("Ignoring invalid flow rule", "rule", rule, "err", err)
			continue
		}
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*Rule, 0, 1)
		}
		resRules = append(resRules, rule)
		resRulesMap[rule.Resource] = resRules
	}

	for res, rulesOfRes := range resRulesMap {
		m[res] = buildRulesOfRes(res, rulesOfRes)
	}
	return m
}

// buildRulesOfRes builds TrafficShapingController slice from rules. the resource of rules must be equals to res
func buildRulesOfRes(res string, rulesOfRes []*Rule) []*TrafficShapingController {
	resTcs := make([]*TrafficShapingController, 0, len(rulesOfRes))
	for _, rule := range rulesOfRes {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name, expect: %s, actual: %s", res, rule.Resource), "FlowManager: unmatched resource name ", "rule", rule)
			continue
		}
		generator, supported := tcGenFuncMap[trafficControllerGenKey{
			tokenCalculateStrategy: rule.TokenCalculateStrategy,
			controlBehavior:        rule.ControlBehavior,
		}]
		if !supported || generator == nil {
			logging.Warn("Ignoring the rule due to unsupported control behavior", "rule", rule)
			continue
		}
		tsc := generator(rule)
		if tsc == nil {
			logging.Warn("Ignoring the rule due to bad generated traffic controller.", "rule", rule)
			continue
		}
		resTcs = append(resTcs, tsc)
	}
	return resTcs
}

// IsValidRule checks whether the given Rule is valid.
func IsValidRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil Rule")
	}
	if rule.Resource == "" {
		return errors.New("empty resource name")
	}
	if rule.Count < 0 {
		return errors.New("negative threshold")
	}
	if rule.MetricType < 0 {
		return errors.New("invalid metric type")
	}
	if rule.RelationStrategy < 0 {
		return errors.New("invalid relation strategy")
	}
	if rule.TokenCalculateStrategy < 0 || rule.ControlBehavior < 0 {
		return errors.New("invalid control strategy")
	}

	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return errors.New("Bad flow rule: invalid control behavior")
	}

	if rule.TokenCalculateStrategy == WarmUp {
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("invalid warmUpPeriodSec")
		}
		if rule.WarmUpColdFactor == 1 {
			return errors.New("WarmUpColdFactor must be great than 1")
		}
	}
	return nil
}
