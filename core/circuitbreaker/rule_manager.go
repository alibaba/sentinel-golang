package circuitbreaker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

type CircuitBreakerGenFunc func(r Rule, reuseStat interface{}) CircuitBreaker

var (
	logger       = logging.GetDefaultLogger()
	cbGenFuncMap = make(map[Strategy]CircuitBreakerGenFunc)

	breakerRules = make(map[string][]Rule)
	breakers     = make(map[string][]CircuitBreaker)
	updateMux    = &sync.RWMutex{}

	statusSwitchListeners = make([]StateChangeListener, 0)
)

func init() {
	cbGenFuncMap[SlowRt] = func(r Rule, reuseStat interface{}) CircuitBreaker {
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

// Load the newer rules to manager.
// rules: the newer rules, if len of rules is 0, will clear all rules of manager.
// return value:
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
			logger.Warnf("Ignoring invalid breaker rule when loading new rules,rule: %+v, err: %+v.", rule, err)
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

	updateMux.Lock()
	defer updateMux.Unlock()

	for res, resRules := range newBreakerRules {
		emptyCircuitBreakerList := make([]CircuitBreaker, 0, 0)
		for _, r := range resRules {
			oldResCbs := breakers[res]
			if oldResCbs == nil {
				oldResCbs = emptyCircuitBreakerList
			}
			equalsIdx := -1
			reuseStatIdx := -1
			for idx, cb := range oldResCbs {
				oldRule := cb.BoundRule()
				if oldRule.IsEqualsTo(r) {
					equalsIdx = idx
					break
				}
				if !oldRule.IsStatReusable(r) {
					continue
				}
				if reuseStatIdx >= 0 {
					// had find reuse rule.
					continue
				}
				reuseStatIdx = idx
			}

			// First check equals scenario
			if equalsIdx >= 0 {
				// reuse the old cb
				reuseOldCb := oldResCbs[equalsIdx]
				cbsOfRes, ok := newBreakers[res]
				if !ok {
					cbsOfRes = make([]CircuitBreaker, 0, 1)
					newBreakers[res] = append(cbsOfRes, reuseOldCb)
				} else {
					newBreakers[res] = append(cbsOfRes, reuseOldCb)
				}
				// remove old cb from oldResCbs
				oldResCbs = append(oldResCbs[:equalsIdx], oldResCbs[equalsIdx+1:]...)
				breakers[res] = oldResCbs
				continue
			}

			generator := cbGenFuncMap[r.BreakerStrategy()]
			if generator == nil {
				logger.Warnf("Circuit Breaker Generator for %+resRules is not existed.", r.BreakerStrategy())
				continue
			}

			var cb CircuitBreaker
			if reuseStatIdx >= 0 {
				cb = generator(r, oldResCbs[reuseStatIdx].BoundStat())
			} else {
				cb = generator(r, nil)
			}
			if cb == nil {
				logger.Warnf("Fail to generate Circuit Breaker for rule: %+resRules.", r)
				continue
			}

			if reuseStatIdx >= 0 {
				breakers[res] = append(oldResCbs[:reuseStatIdx], oldResCbs[reuseStatIdx+1:]...)
			}
			cbsOfRes, ok := newBreakers[res]
			if !ok {
				cbsOfRes = make([]CircuitBreaker, 0, 1)
				newBreakers[res] = append(cbsOfRes, cb)
			} else {
				newBreakers[res] = append(cbsOfRes, cb)
			}
		}
	}

	breakerRules = newBreakerRules
	breakers = newBreakers

	logRuleUpdate(newBreakerRules)
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
	sb.WriteString("[CircuitBreakerRuleManager] succeed to load circuit breakers:\n")
	for _, rule := range rulesFrom(rules) {
		sb.WriteString(rule.String())
		sb.WriteString("\n")
	}
	logger.Info(sb.String())
}

func RegisterStatusSwitchListeners(listeners ...StateChangeListener) {
	if len(listeners) == 0 {
		return
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	statusSwitchListeners = append(statusSwitchListeners, listeners...)
}

// SetTrafficShapingGenerator sets the traffic controller generator for the given control behavior.
// Note that modifying the generator of default control behaviors is not allowed.
func SetTrafficShapingGenerator(s Strategy, generator CircuitBreakerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if s >= SlowRt && s <= ErrorCount {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	cbGenFuncMap[s] = generator
	return nil
}

func RemoveTrafficShapingGenerator(s Strategy) error {
	if s >= SlowRt && s <= ErrorCount {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	updateMux.Lock()
	defer updateMux.Unlock()

	delete(cbGenFuncMap, s)
	return nil
}
