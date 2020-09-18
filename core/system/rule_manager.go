package system

import (
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

type RuleMap map[MetricType][]*Rule

// const
var (
	ruleMap    = make(RuleMap)
	ruleMapMux = new(sync.RWMutex)
)

// GetRules returns all the rules based on copy.
// It doesn't take effect for system module if user changes the rule.
// GetRules need to compete system module's global lock and the high performance losses of copy,
// 		reduce or do not call GetRules if possible
func GetRules() []Rule {
	rules := make([]*Rule, 0, len(ruleMap))
	ruleMapMux.RLock()
	for _, rs := range ruleMap {
		rules = append(rules, rs...)
	}
	ruleMapMux.RUnlock()

	ret := make([]Rule, 0, len(rules))
	for _, r := range rules {
		ret = append(ret, *r)
	}
	return ret
}

// getRules returns all the rules。Any changes of rules take effect for system module
// getRules is an internal interface.
func getRules() []*Rule {
	ruleMapMux.RLock()
	defer ruleMapMux.RUnlock()

	rules := make([]*Rule, 0)
	for _, rs := range ruleMap {
		rules = append(rules, rs...)
	}
	return rules
}

// LoadRules loads given system rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*Rule) (bool, error) {
	m := buildRuleMap(rules)

	if err := onRuleUpdate(m); err != nil {
		logging.Error(err, "Fail to load rules", "rules", rules)
		return false, err
	}

	return true, nil
}

// ClearRules clear all the previous rules
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

func onRuleUpdate(r RuleMap) error {
	start := util.CurrentTimeNano()
	ruleMapMux.Lock()
	defer func() {
		ruleMapMux.Unlock()
		logging.Debug("time statistic(ns) for updating system rule", "timeCost", util.CurrentTimeNano()-start)
		if len(r) > 0 {
			logging.Info("[SystemRuleManager] System rules loaded", "rules", r)
		} else {
			logging.Info("[SystemRuleManager] System rules were cleared")
		}
	}()
	ruleMap = r
	return nil
}

func buildRuleMap(rules []*Rule) RuleMap {
	m := make(RuleMap)

	if len(rules) == 0 {
		return m
	}

	for _, rule := range rules {
		if err := IsValidSystemRule(rule); err != nil {
			logging.Warn("Ignoring invalid system rule", "rule", rule, "err", err)
			continue
		}
		rulesOfRes, exists := m[rule.MetricType]
		if !exists {
			m[rule.MetricType] = []*Rule{rule}
		} else {
			m[rule.MetricType] = append(rulesOfRes, rule)
		}
	}
	return m
}

// IsValidSystemRule determine the system rule is valid or not
func IsValidSystemRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil Rule")
	}
	if rule.TriggerCount < 0 {
		return errors.New("negative threshold")
	}
	if rule.MetricType >= MetricTypeSize {
		return errors.New("invalid metric type")
	}

	if rule.MetricType == CpuUsage && rule.TriggerCount > 1 {
		return errors.New("invalid CPU usage, valid range is [0.0, 1.0]")
	}
	return nil
}
