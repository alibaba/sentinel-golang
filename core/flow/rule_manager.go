package flow

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
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
			logging.Infof("[FlowRuleManager] Flow rules were loaded: %s", bs)
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

	m := buildFlowMap(rules)

	start := util.CurrentTimeNano()
	tcMux.Lock()
	defer func() {
		tcMux.Unlock()
		if r := recover(); r != nil {
			return
		}
		logging.Debugf("Updating flow rule spends %d ns.", util.CurrentTimeNano()-start)
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

// getRules returns all the rules。Any changes of rules take effect for flow module
// getRules is an internal interface.
func getRules() []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return rulesFrom(tcMap)
}

// getResRules returns specific resource's rules。Any changes of rules take effect for flow module
// getResRules is an internal interface.
func getResRules(res string) []*Rule {
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

// GetResRules returns specific resource's rules based on copy.
// It doesn't take effect for flow module if user changes the rule.
func GetResRules(res string) []Rule {
	rules := getResRules(res)
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
func buildFlowMap(rules []*Rule) TrafficControllerMap {
	m := make(TrafficControllerMap)
	if len(rules) == 0 {
		return m
	}

	for _, rule := range rules {
		if err := IsValidRule(rule); err != nil {
			logging.Warnf("Ignoring invalid flow rule: %v, reason: %s", rule, err.Error())
			continue
		}
		generator, supported := tcGenFuncMap[trafficControllerGenKey{
			tokenCalculateStrategy: rule.TokenCalculateStrategy,
			controlBehavior:        rule.ControlBehavior,
		}]
		if !supported {
			logging.Warnf("Ignoring the rule due to unsupported control behavior: %v", rule)
			continue
		}
		tsc := generator(rule)
		if tsc == nil {
			logging.Warnf("Ignoring the rule due to bad generated traffic controller: %v", rule)
			continue
		}

		rulesOfRes, exists := m[rule.Resource]
		if !checkRuleIdCompliance(rulesOfRes, rule) {
			logging.Warnf("repeat rule id: %d", rule.ID)
			continue
		}
		if !exists {
			m[rule.Resource] = []*TrafficShapingController{tsc}
		} else {
			m[rule.Resource] = append(rulesOfRes, tsc)
		}

	}
	return m
}

func checkRuleIdCompliance(resRules []*TrafficShapingController, rule *Rule) bool {
	if rule == nil || len(resRules) == 0 {
		return true
	}
	for _, r := range resRules {
		if r.Rule().ID == 0 || rule.ID == 0 {
			continue
		}
		if r.Rule().ID == rule.ID {
			return false
		}
	}
	return true
}

func UpdateRule(rule *Rule) error {
	if err := IsValidRule(rule); err != nil {
		return err
	}
	if rule.ID == 0 {
		return errors.New("Update rule's id can't be 0")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	resTcs, exists := tcMap[rule.Resource]
	if !exists {
		return errors.New("Can't update not existed rule")
	}

	updateRuleExist := false
	for idx, tc := range resTcs {
		if tc.Rule().ID == rule.ID {
			generator, supported := tcGenFuncMap[trafficControllerGenKey{
				tokenCalculateStrategy: rule.TokenCalculateStrategy,
				controlBehavior:        rule.ControlBehavior,
			}]
			if !supported {
				return errors.New("Ignoring the rule due to unsupported control behavior")
			}
			tsc := generator(rule)
			if tsc == nil {
				return errors.New("Ignoring the rule due to bad generated traffic controller")
			}
			resTcs[idx] = tsc
			tcMap[rule.Resource] = resTcs
			updateRuleExist = true
			break
		}
	}

	if !updateRuleExist {
		return errors.New(fmt.Sprintf("No rule to update was found based on rule id:%d", rule.ID))
	}
	return nil
}

func AppendRule(rule *Rule) error {
	if err := IsValidRule(rule); err != nil {
		return err
	}
	// ruleId is unique in level <resource, module>
	appendRuleId := rule.ID
	tcMux.Lock()
	defer tcMux.Unlock()

	// check rule id availability
	resTcs, ok := tcMap[rule.Resource]
	if ok {
		for _, tc := range resTcs {
			if tc.Rule().ID > 0 && tc.Rule().ID == appendRuleId {
				return errors.New("Valid rule id existed.")
			}
			if isEquivalentRule(rule, tc.Rule()) {
				return errors.New("Equivalent Rule existed.")
			}
		}
	}

	generator, supported := tcGenFuncMap[trafficControllerGenKey{
		tokenCalculateStrategy: rule.TokenCalculateStrategy,
		controlBehavior:        rule.ControlBehavior,
	}]
	if !supported {
		return errors.New("Unsupported control strategy")
	}
	tsc := generator(rule)
	if tsc == nil {
		return errors.New("Bad generated traffic controller")
	}
	if !ok {
		tcMap[rule.Resource] = []*TrafficShapingController{tsc}
	} else {
		tcMap[rule.Resource] = append(resTcs, tsc)
	}
	return nil
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

func isEquivalentRule(rule1 *Rule, rule2 *Rule) bool {
	if !(rule1.Resource == rule2.Resource && rule1.TokenCalculateStrategy == rule2.TokenCalculateStrategy &&
		rule1.ControlBehavior == rule2.ControlBehavior && rule1.RelationStrategy == rule2.RelationStrategy &&
		rule1.RefResource == rule2.RefResource && rule1.MetricType == rule2.MetricType &&
		util.Float64Equals(rule1.Count, rule2.Count)) {
		return false
	}
	if rule1.TokenCalculateStrategy == WarmUp && !(rule1.WarmUpPeriodSec == rule2.WarmUpPeriodSec &&
		rule1.WarmUpColdFactor == rule2.WarmUpColdFactor) {
		return false
	}

	if rule1.ControlBehavior == Throttling && !(rule1.MaxQueueingTimeMs == rule2.MaxQueueingTimeMs) {
		return false
	}

	return true
}
