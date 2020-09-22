package isolation

import (
	"encoding/json"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

var (
	ruleMap = make(map[string][]*Rule)
	rwMux   = &sync.RWMutex{}
)

func LoadRules(rules []*Rule) (updated bool, err error) {
	updated = true
	err = nil

	m := make(map[string][]*Rule)
	for _, r := range rules {
		if e := IsValid(r); e != nil {
			logging.Error(e, "invalid isolation rule.", "rule", r)
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
		logging.Debug("time statistic(ns) for updating isolation rule", "timeCost", util.CurrentTimeNano()-start)
		logRuleUpdate(m)
	}()
	ruleMap = m
	return
}

func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func GetRulesOfResource(res string) []Rule {
	rules := getRulesOfResource(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func getRules() []*Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()

	return rulesFrom(ruleMap)
}

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
	rules := make([]*Rule, 0)
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
	bs, err := json.Marshal(rulesFrom(m))
	if err != nil {
		if len(m) == 0 {
			logging.Info("[IsolationRuleManager] Isolation rules were cleared")
		} else {
			logging.Info("[IsolationRuleManager] Isolation rules were loaded")
		}
	} else {
		if len(m) == 0 {
			logging.Info("[IsolationRuleManager] Isolation rules were cleared")
		} else {
			logging.Info("[IsolationRuleManager] Isolation rules were loaded", "rules", string(bs))
		}
	}
}

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
