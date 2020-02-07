package flow

import (
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/logging"
	"sync"
)

// const
var (
	logger = logging.GetDefaultLogger()
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(*FlowRule) *TrafficShapingController

// TrafficControllerMap represents the map storage for TrafficShapingController.
type TrafficControllerMap map[string][]*TrafficShapingController

var (
	tcGenFuncMap = make(map[ControlBehavior]TrafficControllerGenFunc)
	tcMap        = make(TrafficControllerMap, 0)
	tcMux        = new(sync.RWMutex)

	ruleChan     = make(chan []*FlowRule, 10)
	propertyInit sync.Once
)

func init() {
	propertyInit.Do(func() {
		initRuleRecvTask()
	})

	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[Reject] = func(rule *FlowRule) *TrafficShapingController {
		return NewTrafficShapingController(NewDefaultTrafficShapingCalculator(rule.Count), NewDefaultTrafficShapingChecker(rule.MetricType), rule)
	}
	tcGenFuncMap[Throttling] = func(rule *FlowRule) *TrafficShapingController {
		return NewTrafficShapingController(NewDefaultTrafficShapingCalculator(rule.Count), NewThrottlingChecker(rule.MaxQueueingTimeMs, rule.Count), rule)
	}
}

func initRuleRecvTask() {
	go func() {
		for {
			select {
			case rules := <-ruleChan:
				err := onRuleUpdate(rules)
				if err != nil {
					logger.Errorf("Failed to update flow rules: %+v", err)
				}
			}
		}
	}()
}

func onRuleUpdate(rules []*FlowRule) error {
	tcMux.Lock()
	defer tcMux.Unlock()

	m := buildFlowMap(rules)

	tcMap = m
	return nil
}

// LoadRules loads the given flow rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*FlowRule) (bool, error) {
	// TODO: rethink the design
	ruleChan <- rules
	return true, nil
}

// SetTrafficShapingGenerator sets the traffic controller generator for the given control behavior.
// Note that modifying the generator of default control behaviors is not allowed.
func SetTrafficShapingGenerator(cb ControlBehavior, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if cb >= Reject && cb <= WarmUpThrottling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[cb] = generator
	return nil
}

func RemoveTrafficShapingGenerator(cb ControlBehavior) error {
	if cb >= Reject && cb <= WarmUpThrottling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, cb)
	return nil
}

func getTrafficControllerListFor(name string) []*TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[name]
}

// NotThreadSafe
func buildFlowMap(rules []*FlowRule) TrafficControllerMap {
	if len(rules) == 0 {
		return make(TrafficControllerMap, 0)
	}
	m := make(TrafficControllerMap, 0)
	for _, rule := range rules {
		if !IsValidFlowRule(rule) {
			logger.Warnf("Ignoring invalid flow rule: %v", rule)
			continue
		}
		if rule.LimitOrigin == "" {
			rule.LimitOrigin = LimitOriginDefault
		}
		generator, supported := tcGenFuncMap[rule.ControlBehavior]
		if !supported {
			logger.Warnf("Ignoring the rule due to unsupported control behavior: %v", rule)
			continue
		}
		tsc := generator(rule)

		rulesOfRes, exists := m[rule.Resource]
		if !exists {
			m[rule.Resource] = []*TrafficShapingController{tsc}
		} else {
			m[rule.Resource] = append(rulesOfRes, tsc)
		}
	}
	return m
}

// IsValidFlowRule checks whether the given FlowRule is valid.
func IsValidFlowRule(rule *FlowRule) bool {
	if rule == nil || rule.Resource == "" || rule.Count < 0 {
		return false
	}
	if rule.MetricType < 0 || rule.RelationStrategy < 0 || rule.ControlBehavior < 0 {
		return false
	}

	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return false
	}

	return checkClusterField(rule) && checkControlBehaviorField(rule)
}

func checkClusterField(rule *FlowRule) bool {
	if rule.ClusterMode && rule.ID <= 0 {
		return false
	}
	return true
}

func checkControlBehaviorField(rule *FlowRule) bool {
	switch rule.ControlBehavior {
	case WarmUp:
		return rule.WarmUpPeriodSec > 0
	case Throttling:
		return true
	case WarmUpThrottling:
		return rule.WarmUpPeriodSec > 0 && rule.MaxQueueingTimeMs > 0
	default:
		return true
	}
}
