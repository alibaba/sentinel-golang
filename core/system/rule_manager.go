package system

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
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
	go util.RunWithRecover(func() {
		for {
			select {
			case rules := <-ruleChan:
				err := onRuleUpdate(rules)
				if err != nil {
					logger.Errorf("Failed to update system rules: %+v", err)
				}
			}
		}
	}, logger)
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
	logRuleUpdate(m)

	return nil
}

func logRuleUpdate(m RuleMap) {
	rules := make([]*SystemRule, 0)
	for _, rs := range m {
		rules = append(rules, rs...)
	}
	bs, err := json.Marshal(rules)
	if err != nil {
		logger.Info("[SystemRuleManager] System rules loaded")
	} else {
		logger.Infof("[SystemRuleManager] System rules loaded: %s", bs)
	}
}

func buildRuleMap(rules []*SystemRule) RuleMap {
	if len(rules) == 0 {
		return make(RuleMap, 0)
	}
	m := make(RuleMap, 0)
	for _, rule := range rules {
		if err := IsValidSystemRule(rule); err != nil {
			logger.Warnf("Ignoring invalid system rule: %v, reason: %s", rule, err.Error())
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

func IsValidSystemRule(rule *SystemRule) error {
	if rule == nil {
		return errors.New("nil SystemRule")
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
