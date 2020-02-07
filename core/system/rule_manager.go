package system

import (
	"github.com/sentinel-group/sentinel-golang/logging"
	"sync"
)

type RuleMap map[MetricType][]*SystemRule

// const
var (
	logger = logging.GetDefaultLogger()

	ruleMap    = make(RuleMap, 0)
	ruleMapMux = new(sync.RWMutex)

	ruleChan     = make(chan []*SystemRule, 10)
	propertyInit sync.Once
)

func init() {
	propertyInit.Do(func() {
		initRuleRecvTask()
	})
}

func initRuleRecvTask() {
	go func() {
		for {
			select {
			case rules := <-ruleChan:
				err := onRuleUpdate(rules)
				if err != nil {
					logger.Errorf("Failed to update system rules: %+v", err)
				}
			}
		}
	}()
}

func GetRules() []*SystemRule {
	ruleMapMux.RLock()
	defer ruleMapMux.RUnlock()

	rules := make([]*SystemRule, 0)
	for _, rs := range ruleMap {
		rules = append(rules, rs...)
	}
	return rules
}

// LoadRules loads given system rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*SystemRule) (bool, error) {
	ruleChan <- rules
	return true, nil
}

func onRuleUpdate(rules []*SystemRule) error {
	m := buildRuleMap(rules)

	ruleMapMux.Lock()
	defer ruleMapMux.Unlock()

	ruleMap = m
	return nil
}

func buildRuleMap(rules []*SystemRule) RuleMap {
	if len(rules) == 0 {
		return make(RuleMap, 0)
	}
	m := make(RuleMap, 0)
	for _, rule := range rules {
		if !IsValidSystemRule(rule) {
			logger.Warnf("Ignoring invalid system rule: %v", rule)
			continue
		}
		rulesOfRes, exists := m[rule.MetricType]
		if !exists {
			m[rule.MetricType] = []*SystemRule{rule}
		} else {
			m[rule.MetricType] = append(rulesOfRes, rule)
		}
	}
	return m
}

func IsValidSystemRule(rule *SystemRule) bool {
	if rule == nil || rule.TriggerCount < 0 || rule.MetricType >= MetricTypeSize {
		return false
	}
	if rule.MetricType == CpuUsage && rule.TriggerCount > 1 {
		return false
	}
	return true
}
