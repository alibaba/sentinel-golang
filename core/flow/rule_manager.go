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

// TrafficControllerMap represents the map storage for TrafficShapingController.
type TrafficControllerMap map[string][]*TrafficShapingController

var (
	tcGenFuncMap = make(map[ControlStrategy]TrafficControllerGenFunc)
	tcMap        = make(TrafficControllerMap)
	tcMux        = new(sync.RWMutex)
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[ControlStrategy{
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewDirectTrafficShapingCalculator(rule.Count), NewDefaultTrafficShapingChecker(rule), rule)
	}
	tcGenFuncMap[ControlStrategy{
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Throttling,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewDirectTrafficShapingCalculator(rule.Count), NewThrottlingChecker(rule.MaxQueueingTimeMs), rule)
	}
	tcGenFuncMap[ControlStrategy{
		TokenCalculateStrategy: WarmUp,
		ControlBehavior:        Reject,
	}] = func(rule *Rule) *TrafficShapingController {
		return NewTrafficShapingController(NewWarmUpTrafficShapingCalculator(rule), NewDefaultTrafficShapingChecker(rule), rule)
	}
	tcGenFuncMap[ControlStrategy{
		TokenCalculateStrategy: WarmUp,
		ControlBehavior:        Throttling,
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

func getRules() []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return rulesFrom(tcMap)
}

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

func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func GetResRules(res string) []Rule {
	rules := getResRules(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

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

// SetTrafficShapingGenerator sets the traffic controller generator for the given control strategy.
// Note that modifying the generator of default control strategy is not allowed.
func SetTrafficShapingGenerator(cs ControlStrategy, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if cs.TokenCalculateStrategy >= Direct && cs.TokenCalculateStrategy <= WarmUp {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	if cs.ControlBehavior >= Reject && cs.ControlBehavior <= Throttling {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[cs] = generator
	return nil
}

func RemoveTrafficShapingGenerator(cs ControlStrategy) error {
	if cs.TokenCalculateStrategy >= Direct && cs.TokenCalculateStrategy <= WarmUp {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	if cs.ControlBehavior >= Reject && cs.ControlBehavior <= Throttling {
		return errors.New("not allowed to replace the generator for default control strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, cs)
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
		generator, supported := tcGenFuncMap[rule.ControlStrategy]
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
		if !exists {
			m[rule.Resource] = []*TrafficShapingController{tsc}
		} else {
			m[rule.Resource] = append(rulesOfRes, tsc)
		}
	}
	return m
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
	if rule.ControlStrategy.TokenCalculateStrategy < 0 || rule.ControlStrategy.ControlBehavior < 0 {
		return errors.New("invalid control strategy")
	}

	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return errors.New("Bad flow rule: invalid control behavior")
	}

	return checkControlStrategyField(rule)
}

func checkControlStrategyField(rule *Rule) error {
	switch rule.ControlStrategy.TokenCalculateStrategy {
	case WarmUp:
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("invalid warmUpPeriodSec")
		}
		if rule.WarmUpColdFactor == 1 {
			return errors.New("WarmUpColdFactor must be great than 1")
		}
		return nil
	default:
	}
	return nil
}
