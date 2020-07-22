package circuitbreaker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

type CircuitBreakerGenFunc func(r Rule, reuseStat interface{}) CircuitBreaker

var (
	logger       = logging.GetDefaultLogger()
	cbGenFuncMap = make(map[Strategy]CircuitBreakerGenFunc)

	breakerRules = make(map[string][]Rule)
	breakers     = make(map[string][]CircuitBreaker)
	updateMux    = &sync.RWMutex{}

	stateChangeListeners = make([]StateChangeListener, 0)
)

func init() {
	cbGenFuncMap[SlowRequestRatio] = func(r Rule, reuseStat interface{}) CircuitBreaker {
		rtRule, ok := r.(*slowRtRule)
		if !ok || rtRule == nil {
			return nil
		}
		if reuseStat == nil {
			return newSlowRtCircuitBreaker(rtRule)
		}
		stat, ok := reuseStat.(*slowRequestLeapArray)
		if !ok || stat == nil {
			logger.Warnf("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*slowRequestLeapArray, in fact: %+v", stat)
			return newSlowRtCircuitBreaker(rtRule)
		}
		return newSlowRtCircuitBreakerWithStat(rtRule, stat)
	}

	cbGenFuncMap[ErrorRatio] = func(r Rule, reuseStat interface{}) CircuitBreaker {
		errRatioRule, ok := r.(*errorRatioRule)
		if !ok || errRatioRule == nil {
			return nil
		}
		if reuseStat == nil {
			return newErrorRatioCircuitBreaker(errRatioRule)
		}
		stat, ok := reuseStat.(*errorCounterLeapArray)
		if !ok || stat == nil {
			logger.Warnf("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*errorCounterLeapArray, in fact: %+v", stat)
			return newErrorRatioCircuitBreaker(errRatioRule)
		}
		return newErrorRatioCircuitBreakerWithStat(errRatioRule, stat)
	}

	cbGenFuncMap[ErrorCount] = func(r Rule, reuseStat interface{}) CircuitBreaker {
		errCountRule, ok := r.(*errorCountRule)
		if !ok || errCountRule == nil {
			return nil
		}
		if reuseStat == nil {
			return newErrorCountCircuitBreaker(errCountRule)
		}
		stat, ok := reuseStat.(*errorCounterLeapArray)
		if !ok || stat == nil {
			logger.Warnf("Expect to generate circuit breaker with reuse statistic, but fail to do type assertion, expect:*errorCounterLeapArray, in fact: %+v", stat)
			return newErrorCountCircuitBreaker(errCountRule)
		}
		return newErrorCountCircuitBreakerWithStat(errCountRule, stat)
	}
}

func GetResRules(resource string) []Rule {
	updateMux.RLock()
	ret, ok := breakerRules[resource]
	updateMux.RUnlock()
	if !ok {
		ret = make([]Rule, 0)
	}
	return ret
}

// ClearRules clear all the previous rules.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// LoadRules replaces old rules with the given circuit breaking rules.
//
// return value:
//
// bool: was designed to indicate whether the internal map has been changed
// error: was designed to indicate whether occurs the error.
func LoadRules(rules []Rule) (bool, error) {
	// TODO in order to avoid invalid update, should check consistent with last update rules
	err := onRuleUpdate(rules)
	return true, err
}

func getResBreakers(resource string) []CircuitBreaker {
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

func calculateReuseIndexFor(r Rule, oldResCbs []CircuitBreaker) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old circuit breaker slice
	equalIdx = -1
	// the index of statistic reusable rule in old circuit breaker slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResCbs {
		oldRule := oldTc.BoundRule()
		if oldRule.IsEqualsTo(r) {
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
func onRuleUpdate(rules []Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	newBreakerRules := make(map[string][]Rule)
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if err := rule.IsApplicable(); err != nil {
			logger.Warnf("Ignoring invalid circuit breaking rule when loading new rules, rule: %+v, reason: %s", rule, err.Error())
			continue
		}

		classification := rule.ResourceName()
		ruleSet, ok := newBreakerRules[classification]
		if !ok {
			ruleSet = make([]Rule, 0, 1)
		}
		ruleSet = append(ruleSet, rule)
		newBreakerRules[classification] = ruleSet
	}

	newBreakers := make(map[string][]CircuitBreaker)
	// in order to avoid growing, build newBreakers in advance
	for res, rules := range newBreakerRules {
		newBreakers[res] = make([]CircuitBreaker, 0, len(rules))
	}

	start := util.CurrentTimeNano()
	updateMux.Lock()
	defer func() {
		updateMux.Unlock()
		if r := recover(); r != nil {
			return
		}
		logger.Debugf("Updating circuit breaker rule spends %d ns.", util.CurrentTimeNano()-start)
		logRuleUpdate(newBreakerRules)
	}()

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
				// reuse the old cb
				equalOldCb := oldResCbs[equalIdx]
				insertCbToCbMap(equalOldCb, res, newBreakers)
				// remove old cb from oldResCbs
				breakers[res] = append(oldResCbs[:equalIdx], oldResCbs[equalIdx+1:]...)
				continue
			}

			generator := cbGenFuncMap[r.BreakerStrategy()]
			if generator == nil {
				logger.Warnf("Ignoring the rule due to unsupported circuit breaking strategy: %v", r)
				continue
			}

			var cb CircuitBreaker
			if reuseStatIdx >= 0 {
				cb = generator(r, oldResCbs[reuseStatIdx].BoundStat())
			} else {
				cb = generator(r, nil)
			}
			if cb == nil {
				logger.Warnf("Ignoring the rule due to bad generated circuit breaker: %v", r)
				continue
			}

			if reuseStatIdx >= 0 {
				breakers[res] = append(oldResCbs[:reuseStatIdx], oldResCbs[reuseStatIdx+1:]...)
			}
			insertCbToCbMap(cb, res, newBreakers)
		}
	}

	breakerRules = newBreakerRules
	breakers = newBreakers
	return nil
}

func rulesFrom(rm map[string][]Rule) []Rule {
	rules := make([]Rule, 0)
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

func logRuleUpdate(rules map[string][]Rule) {
	sb := strings.Builder{}
	sb.WriteString("Circuit breaking rules loaded: [")

	for _, r := range rulesFrom(rules) {
		sb.WriteString(r.String() + ",")
	}
	sb.WriteString("]")
	logger.Info(sb.String())
}

func RegisterStateChangeListeners(listeners ...StateChangeListener) {
	if len(listeners) == 0 {
		return
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	stateChangeListeners = append(stateChangeListeners, listeners...)
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
