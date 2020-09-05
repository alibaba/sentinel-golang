package hotspot

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController

// trafficControllerMap represents the map storage for TrafficShapingController.
type trafficControllerMap map[string][]TrafficShapingController

var (
	tcGenFuncMap = make(map[ControlBehavior]TrafficControllerGenFunc)
	tcMap        = make(trafficControllerMap)
	tcMux        = new(sync.RWMutex)
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[Reject] = func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController {
		var baseTc *baseTrafficShapingController
		if reuseMetric != nil {
			// new BaseTrafficShapingController with reuse statistic metric
			baseTc = newBaseTrafficShapingControllerWithMetric(r, reuseMetric)
		} else {
			baseTc = newBaseTrafficShapingController(r)
		}
		return &rejectTrafficShapingController{
			baseTrafficShapingController: *baseTc,
			burstCount:                   r.BurstCount,
		}
	}

	tcGenFuncMap[Throttling] = func(r *Rule, reuseMetric *ParamsMetric) TrafficShapingController {
		var baseTc *baseTrafficShapingController
		if reuseMetric != nil {
			baseTc = newBaseTrafficShapingControllerWithMetric(r, reuseMetric)
		} else {
			baseTc = newBaseTrafficShapingController(r)
		}
		return &throttlingTrafficShapingController{
			baseTrafficShapingController: *baseTc,
			maxQueueingTimeMs:            r.MaxQueueingTimeMs,
		}
	}
}

func getTrafficControllersFor(res string) []TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[res]
}

// LoadRules replaces old rules with the given hotspot parameter flow control rules. Return value:
//
// bool: indicates whether the internal map has been changed;
// error: indicates whether occurs the error.
func LoadRules(rules []*Rule) (bool, error) {
	err := onRuleUpdate(rules)
	return true, err
}

// GetRules returns existing rules of the given resource.
func GetRules(res string) []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()
	resTcs := tcMap[res]
	ret := make([]*Rule, 0, len(resTcs))
	for _, tc := range resTcs {
		ret = append(ret, tc.BoundRule())
	}
	return ret
}

// ClearRules clears all parameter flow rules.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

func onRuleUpdate(rules []*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	newRuleMap := make(map[string][]*Rule)
	for _, r := range rules {
		if err := IsValidRule(r); err != nil {
			logging.Warnf("Ignoring invalid hotspot rule when loading new rules, rule: %s, reason: %s", r.String(), err.Error())
			continue
		}
		res := r.ResourceName()
		ruleSet, ok := newRuleMap[res]
		if !ok {
			ruleSet = make([]*Rule, 0, 1)
		}
		ruleSet = append(ruleSet, r)
		newRuleMap[res] = ruleSet
	}

	m := make(trafficControllerMap)
	for res, rules := range newRuleMap {
		m[res] = make([]TrafficShapingController, 0, len(rules))
	}

	start := util.CurrentTimeNano()
	tcMux.Lock()
	defer func() {
		tcMux.Unlock()
		if r := recover(); r != nil {
			return
		}
		logging.Debugf("Updating hotspot rule spends %d ns.", util.CurrentTimeNano()-start)
		logRuleUpdate(m)
	}()

	for res, resRules := range newRuleMap {
		emptyTcList := make([]TrafficShapingController, 0, 0)
		for _, r := range resRules {
			oldResTcs := tcMap[res]
			if oldResTcs == nil {
				oldResTcs = emptyTcList
			}

			equalIdx, reuseStatIdx := calculateReuseIndexFor(r, oldResTcs)
			// there is equivalent rule in old traffic shaping controller slice
			if equalIdx >= 0 {
				equalOldTC := oldResTcs[equalIdx]
				insertTcToTcMap(equalOldTC, res, m)
				// remove old tc from old resTcs
				tcMap[res] = append(oldResTcs[:equalIdx], oldResTcs[equalIdx+1:]...)
				continue
			}

			// generate new traffic shaping controller
			generator, supported := tcGenFuncMap[r.ControlBehavior]
			if !supported {
				logging.Warnf("Ignoring the frequent param flow rule due to unsupported control behavior: %v", r)
				continue
			}
			var tc TrafficShapingController
			if reuseStatIdx >= 0 {
				// generate new traffic shaping controller with reusable statistic metric.
				tc = generator(r, oldResTcs[reuseStatIdx].BoundMetric())
			} else {
				tc = generator(r, nil)
			}
			if tc == nil {
				logging.Debugf("Ignoring the frequent param flow rule due to bad generated traffic controller: %v", r)
				continue
			}

			//  remove the reused traffic shaping controller old res tcs
			if reuseStatIdx >= 0 {
				tcMap[res] = append(oldResTcs[:reuseStatIdx], oldResTcs[reuseStatIdx+1:]...)
			}
			insertTcToTcMap(tc, res, m)
		}
	}
	tcMap = m

	return nil
}

func logRuleUpdate(m trafficControllerMap) {
	sb := strings.Builder{}
	sb.WriteString("Hotspot parameter flow control rules loaded: [")

	for _, r := range rulesFrom(m) {
		sb.WriteString(r.String() + ",")
	}
	sb.WriteString("]")
	logging.Info(sb.String())
}

func rulesFrom(m trafficControllerMap) []*Rule {
	rules := make([]*Rule, 0)
	if len(m) == 0 {
		return rules
	}
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		for _, r := range rs {
			if r != nil && r.BoundRule() != nil {
				rules = append(rules, r.BoundRule())
			}
		}
	}
	return rules
}

func calculateReuseIndexFor(r *Rule, oldResTcs []TrafficShapingController) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old traffic shaping controller slice
	equalIdx = -1
	// the index of statistic reusable rule in old traffic shaping controller slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResTcs {
		oldRule := oldTc.BoundRule()
		if oldRule.Equals(r) {
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

func insertTcToTcMap(tc TrafficShapingController, res string, m trafficControllerMap) {
	tcsOfRes, exists := m[res]
	if !exists {
		tcsOfRes = make([]TrafficShapingController, 0, 1)
		m[res] = append(tcsOfRes, tc)
	} else {
		m[res] = append(tcsOfRes, tc)
	}
}

func IsValidRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil hotspot Rule")
	}
	if len(rule.Resource) == 0 {
		return errors.New("empty resource name")
	}
	if rule.Threshold < 0 {
		return errors.New("negative threshold")
	}
	if rule.MetricType < 0 {
		return errors.New("invalid metric type")
	}
	if rule.ControlBehavior < 0 {
		return errors.New("invalid control strategy")
	}
	if rule.ParamIndex < 0 {
		return errors.New("invalid param index")
	}
	if rule.DurationInSec < 0 {
		return errors.New("invalid duration")
	}
	return checkControlBehaviorField(rule)
}

func checkControlBehaviorField(rule *Rule) error {
	switch rule.ControlBehavior {
	case Reject:
		if rule.BurstCount < 0 {
			return errors.New("invalid BurstCount")
		}
		return nil
	case Throttling:
		if rule.MaxQueueingTimeMs < 0 {
			return errors.New("invalid MaxQueueingTimeMs")
		}
		return nil
	default:
	}
	return nil
}

// SetTrafficShapingGenerator sets the traffic controller generator for the given control behavior.
// Note that modifying the generator of default control behaviors is not allowed.
func SetTrafficShapingGenerator(cb ControlBehavior, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if cb >= Reject && cb <= Throttling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[cb] = generator
	return nil
}

func RemoveTrafficShapingGenerator(cb ControlBehavior) error {
	if cb >= Reject && cb <= Throttling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, cb)
	return nil
}
