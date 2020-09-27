package circuitbreaker

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

type CircuitBreakerGenFunc func(r *Rule, reuseStat interface{}) (CircuitBreaker, error)

var (
	cbGenFuncMap = make(map[Strategy]CircuitBreakerGenFunc)

	breakerRules = make(map[string][]*Rule)
	breakers     = make(map[string][]CircuitBreaker)
	updateMux    = &sync.RWMutex{}

	stateChangeListeners = make([]StateChangeListener, 0)
)

func init() {
	cbGenFuncMap[SlowRequestRatio] = func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
		if r == nil {
			return nil, errors.New("nil rule")
		}
		if reuseStat == nil {
			return newSlowRtCircuitBreaker(r)
		}
		stat, ok := reuseStat.(*slowRequestLeapArray)
		if !ok || stat == nil {
			logging.Warn("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*slowRequestLeapArray", "statType", reflect.TypeOf(stat).Name())
			return newSlowRtCircuitBreaker(r)
		}
		return newSlowRtCircuitBreakerWithStat(r, stat), nil
	}

	cbGenFuncMap[ErrorRatio] = func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
		if r == nil {
			return nil, errors.New("nil rule")
		}
		if reuseStat == nil {
			return newErrorRatioCircuitBreaker(r)
		}
		stat, ok := reuseStat.(*errorCounterLeapArray)
		if !ok || stat == nil {
			logging.Warn("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*errorCounterLeapArray", "statType", reflect.TypeOf(stat).Name())
			return newErrorRatioCircuitBreaker(r)
		}
		return newErrorRatioCircuitBreakerWithStat(r, stat), nil
	}

	cbGenFuncMap[ErrorCount] = func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
		if r == nil {
			return nil, errors.New("nil rule")
		}
		if reuseStat == nil {
			return newErrorCountCircuitBreaker(r)
		}
		stat, ok := reuseStat.(*errorCounterLeapArray)
		if !ok || stat == nil {
			logging.Warn("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*errorCounterLeapArray", "statType", reflect.TypeOf(stat).Name())
			return newErrorCountCircuitBreaker(r)
		}
		return newErrorCountCircuitBreakerWithStat(r, stat), nil
	}
}

// GetRulesOfResource returns specific resource's rules based on copy.
// It doesn't take effect for circuit breaker module if user changes the rule.
// GetRulesOfResource need to compete circuit breaker module's global lock and the high performance losses of copy,
// 		reduce or do not call GetRulesOfResource frequently if possible
func GetRulesOfResource(resource string) []Rule {
	updateMux.RLock()
	resRules, ok := breakerRules[resource]
	updateMux.RUnlock()
	if !ok {
		return nil
	}
	ret := make([]Rule, 0, len(resRules))
	for _, rule := range resRules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for circuit breaker module if user changes the rule.
// GetRules need to compete circuit breaker module's global lock and the high performance losses of copy,
// 		reduce or do not call GetRules if possible
func GetRules() []Rule {
	updateMux.RLock()
	rules := rulesFrom(breakerRules)
	updateMux.RUnlock()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// ClearRules clear all the previous rules.
func ClearRules() error {
	_, err, _ := LoadRules(nil)
	return err
}

// LoadRules replaces old rules with the given circuit breaking rules.
//
// return value:
//
// bool: was designed to indicate whether the internal map has been changed
// error: was designed to indicate whether occurs the error.
// []*Rule: was designed to return failed rules. If there is an error, it returns input rules.
func LoadRules(rules []*Rule) (bool, error, []*Rule) {
	ret, err, failedRules := onRuleUpdate(rules)
	return ret, err, failedRules
}

func getBreakersOfResource(resource string) []CircuitBreaker {
	ret := make([]CircuitBreaker, 0)
	updateMux.RLock()
	resCBs := breakers[resource]
	updateMux.RUnlock()
	if len(resCBs) == 0 {
		return ret
	}
	ret = append(ret, resCBs...)
	return ret
}

func calculateReuseIndexFor(r *Rule, oldResCbs []CircuitBreaker) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old circuit breaker slice
	equalIdx = -1
	// the index of statistic reusable rule in old circuit breaker slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResCbs {
		oldRule := oldTc.BoundRule()
		if oldRule.equalsTo(r) {
			// break if there is equivalent rule
			equalIdx = idx
			break
		}
		// find the index of first StatReusable rule
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

func insertCbToCbMap(cb CircuitBreaker, res string, m map[string][]CircuitBreaker) {
	cbsOfRes, exists := m[res]
	if !exists {
		cbsOfRes = make([]CircuitBreaker, 0, 1)
		m[res] = append(cbsOfRes, cb)
	} else {
		m[res] = append(cbsOfRes, cb)
	}
}

// Concurrent safe to update rules
func onRuleUpdate(rules []*Rule) (ret bool, err error, failedRules []*Rule) {
	var start uint64
	newBreakerRules := make(map[string][]*Rule)

	defer func() {
		if r := recover(); r != nil {
			// Set to false since rules are not updated due to panic
			var ok bool

			ret = false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
			failedRules = rules
			return
		}
		logging.Debug("Time statistics(ns) for updating circuit breaker rule", "timeCost", util.CurrentTimeNano()-start)
		logRuleUpdate(newBreakerRules)
	}()

	// Preset slice capacity to avoid dynamic allocation
	failedRules = make([]*Rule, 0, len(rules))
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if err := IsValid(rule); err != nil {
			failedRules = append(failedRules, rule)
			logging.Warn("Ignoring invalid circuit breaking rule when loading new rules", "rule", rule, "err", err)
			continue
		}

		classification := rule.ResourceName()
		ruleSet, ok := newBreakerRules[classification]
		if !ok {
			ruleSet = make([]*Rule, 0, 1)
		}

		// Deduplicate loading rules
		for _, cmpRule := range ruleSet {
			if rule.equalsTo(cmpRule) {
				rule = nil
				break
			}
		}
		if rule == nil {
			continue
		}

		ruleSet = append(ruleSet, rule)
		newBreakerRules[classification] = ruleSet
	}

	newBreakers := make(map[string][]CircuitBreaker)
	// Preset slice capacity to avoid dynamic allocation
	for res, rules := range newBreakerRules {
		newBreakers[res] = make([]CircuitBreaker, 0, len(rules))
	}
	toAddBreakerRules := make(map[string][]*Rule)
	for res, rules := range toAddBreakerRules {
		toAddBreakerRules[res] = make([]*Rule, 0, len(rules))
	}

	start = util.CurrentTimeNano()
	updateMux.Lock()
	defer updateMux.Unlock()

	for res, resRules := range newBreakerRules {
		emptyCircuitBreakerList := make([]CircuitBreaker, 0, 0)
		for _, r := range resRules {
			oldResCbs := breakers[res]
			if oldResCbs == nil {
				oldResCbs = emptyCircuitBreakerList
			}
			equalIdx, reuseStatIdx := calculateReuseIndexFor(r, oldResCbs)

			// First check equals scenario
			if equalIdx >= 0 {
				newRuleSet := toAddBreakerRules[res]
				newRuleSet = append(newRuleSet, r)
				toAddBreakerRules[res] = newRuleSet
				// reuse the old cb
				equalOldCb := oldResCbs[equalIdx]
				insertCbToCbMap(equalOldCb, res, newBreakers)
				// remove old cb from oldResCbs
				breakers[res] = append(oldResCbs[:equalIdx], oldResCbs[equalIdx+1:]...)
				continue
			}

			generator := cbGenFuncMap[r.Strategy]
			if generator == nil {
				failedRules = append(failedRules, r)
				logging.Warn("Ignoring the rule due to unsupported circuit breaking strategy", "rule", r)
				continue
			}

			var cb CircuitBreaker
			var e error
			if reuseStatIdx >= 0 {
				cb, e = generator(r, oldResCbs[reuseStatIdx].BoundStat())
			} else {
				cb, e = generator(r, nil)
			}
			if cb == nil || e != nil {
				failedRules = append(failedRules, r)
				logging.Warn("Ignoring the rule due to bad generated circuit breaker", "rule", r, "err", e)
				continue
			}

			newRuleSet, ok := toAddBreakerRules[res]
			if !ok {
				newRuleSet = make([]*Rule, 0, 1)
			}
			newRuleSet = append(newRuleSet, r)
			toAddBreakerRules[res] = newRuleSet
			insertCbToCbMap(cb, res, newBreakers)
			// Set to true since a new rule was just added
			ret = true
		}
	}

	// Instead of adding more new rules, the rule set could also have been reduced to less rules.
	// In this case, we compare the added rules with old ones on their sizes.
	if len(toAddBreakerRules) != len(breakerRules) {
		ret = true
	}
	if ret == false {
		for res, addRules := range toAddBreakerRules {
			originRules, ok := breakerRules[res]
			if !ok || (len(addRules) != len(originRules)) {
				ret = true
				break
			}
		}
	}

	breakerRules = toAddBreakerRules
	breakers = newBreakers
	return
}

func rulesFrom(rm map[string][]*Rule) []*Rule {
	rules := make([]*Rule, 0)
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

func logRuleUpdate(m map[string][]*Rule) {
	rs := rulesFrom(m)
	if len(rs) == 0 {
		logging.Info("[CircuitBreakerRuleManager] Circuit breaking rules were cleared")
	} else {
		logging.Info("[CircuitBreakerRuleManager] Circuit breaking rules were loaded", "rules", rs)
	}
}

// Note: this function is not thread-safe.
func RegisterStateChangeListeners(listeners ...StateChangeListener) {
	if len(listeners) == 0 {
		return
	}

	stateChangeListeners = append(stateChangeListeners, listeners...)
}

// ClearStateChangeListeners will clear the all StateChangeListener
// Note: this function is not thread-safe.
func ClearStateChangeListeners() {
	stateChangeListeners = make([]StateChangeListener, 0)
}

// SetCircuitBreakerGenerator sets the circuit breaker generator for the given strategy.
// Note that modifying the generator of default strategies is not allowed.
func SetCircuitBreakerGenerator(s Strategy, generator CircuitBreakerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if s >= SlowRequestRatio && s <= ErrorCount {
		return errors.New("not allowed to replace the generator for default circuit breaking strategies")
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	cbGenFuncMap[s] = generator
	return nil
}

func RemoveCircuitBreakerGenerator(s Strategy) error {
	if s >= SlowRequestRatio && s <= ErrorCount {
		return errors.New("not allowed to replace the generator for default circuit breaking strategies")
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	delete(cbGenFuncMap, s)
	return nil
}

func IsValid(r *Rule) error {
	if len(r.Resource) == 0 {
		return errors.New("empty resource name")
	}
	if uint(r.Strategy) >= uint(UnsupportedStrategy) {
		return errors.New("invalid Strategy")
	}
	if r.StatIntervalMs <= 0 {
		return errors.New("invalid StatIntervalMs")
	}
	if r.RetryTimeoutMs <= 0 {
		return errors.New("invalid RetryTimeoutMs")
	}
	if r.Threshold < 0.0 {
		return errors.New("invalid Threshold")
	}
	if r.Strategy == SlowRequestRatio && r.Threshold > 1.0 {
		return errors.New("invalid slow request ratio threshold (valid range: [0.0, 1.0])")
	}
	if r.Strategy == ErrorRatio && r.Threshold > 1.0 {
		return errors.New("invalid error ratio threshold (valid range: [0.0, 1.0])")
	}
	return nil
}
