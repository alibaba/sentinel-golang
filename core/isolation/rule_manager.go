package isolation

import (
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/core/misc"
	"github.com/alibaba/sentinel-golang/util"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

var (
	ruleMap      = make(map[string][]*Rule)
	rwMux        = &sync.RWMutex{}
	currentRules = make([]*Rule, 0)
)

// LoadRules loads the given isolation rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*Rule) (bool, error) {
	rwMux.RLock()
	isEqual := reflect.DeepEqual(currentRules, rules)
	rwMux.RUnlock()
	if isEqual {
		logging.Info("[Isolation] Load rules repetition, does not load")
		return false, nil
	}

	err := onRuleUpdate(rules)
	return true, err
}

func onRuleUpdate(rules []*Rule) (err error) {
	m := make(map[string][]*Rule, len(rules))
	for _, r := range rules {
		if e := IsValid(r); e != nil {
			logging.Error(e, "Invalid isolation rule in isolation.LoadRules()", "rule", r)
			continue
		}
		resRules, ok := m[r.Resource]
		if !ok {
			resRules = make([]*Rule, 0, 1)
		}
		m[r.Resource] = append(resRules, r)
	}

	start := util.CurrentTimeNano()
	rwMux.Lock()
	defer func() {
		rwMux.Unlock()
		logging.Debug("[Isolation LoadRules] Time statistic(ns) for updating isolation rule", "timeCost", util.CurrentTimeNano()-start)
		logRuleUpdate(m)
	}()

	for res, rs := range m {
		if len(rs) > 0 {
			// update resource slot chain
			misc.RegisterRuleCheckSlotForResource(res, DefaultSlot)
		}
	}
	ruleMap = m
	currentRules = rules
	return
}

// ClearRules clears all the rules in isolation module.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for isolation module if user changes the rule.
func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource returns specific resource's rules based on copy.
// It doesn't take effect for isolation module if user changes the rule.
func GetRulesOfResource(res string) []Rule {
	rules := getRulesOfResource(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// getRules returns all the rules。Any changes of rules take effect for isolation module
// getRules is an internal interface.
func getRules() []*Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()

	return rulesFrom(ruleMap)
}

// getRulesOfResource returns specific resource's rules。Any changes of rules take effect for isolation module
// getRulesOfResource is an internal interface.
func getRulesOfResource(res string) []*Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()

	resRules, exist := ruleMap[res]
	if !exist {
		return nil
	}
	ret := make([]*Rule, 0, len(resRules))
	for _, r := range resRules {
		ret = append(ret, r)
	}
	return ret
}

func rulesFrom(m map[string][]*Rule) []*Rule {
	rules := make([]*Rule, 0, 8)
	if len(m) == 0 {
		return rules
	}
	for _, rs := range m {
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
		logging.Info("[IsolationRuleManager] Isolation rules were cleared")
	} else {
		logging.Info("[IsolationRuleManager] Isolation rules were loaded", "rules", rs)
	}
}

// IsValidRule checks whether the given Rule is valid.
func IsValid(r *Rule) error {
	if r == nil {
		return errors.New("nil isolation rule")
	}
	if len(r.Resource) == 0 {
		return errors.New("empty resource of isolation rule")
	}
	if r.MetricType != Concurrency {
		return errors.Errorf("unsupported metric type: %d", r.MetricType)
	}
	if r.Threshold == 0 {
		return errors.New("zero threshold")
	}
	return nil
}
