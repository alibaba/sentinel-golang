package flow

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

var (
	logger = logging.GetDefaultLogger()
)

// SafeTrafficControllers represents the threadSafe map storage for TrafficShapingController.
type SafeTrafficControllers struct {
	controller map[string][]*TrafficShapingController
	mux        sync.Mutex
}

// NewSafeTrafficControllers returns a new SafeTrafficControllers instance.
func NewSafeTrafficControllers() *SafeTrafficControllers {
	return &SafeTrafficControllers{
		controller: make(map[string][]*TrafficShapingController),
		mux:        sync.Mutex{},
	}
}

// Store sets the value for a key.
func (s *SafeTrafficControllers) Store(key string, value []*TrafficShapingController) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.controller[key] = value
}

// Delete deletes the value for a key
func (s *SafeTrafficControllers) Delete(key string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.controller, key)
}

// Load returns the value stored in the map for a key.
func (s *SafeTrafficControllers) Load(key string) ([]*TrafficShapingController, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	v, ok := s.controller[key]
	return v, ok
}

// Len returns the length of a map.
func (s *SafeTrafficControllers) Len() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return len(s.controller)
}

// GetRules returns flow rules from the TrafficController.
func (s *SafeTrafficControllers) GetRules() []*FlowRule {
	rules := make([]*FlowRule, 0)
	if s.Len() == 0 {
		return rules
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	for _, rs := range s.controller {
		if len(rs) == 0 {
			continue
		}
		for _, r := range rs {
			if r != nil && r.Rule() != nil {
				rules = append(rules, r.Rule())
			}
		}
	}
	return rules
}

// Reset reset the controller map.
func (s *SafeTrafficControllers) Reset(m map[string][]*TrafficShapingController) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.controller = m
}

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(*FlowRule) *TrafficShapingController

// SafeTrafficGenFuncs represents the threadSafe map storage for TrafficControllerGenFunc.
type SafeTrafficGenFuncs struct {
	funcs map[ControlBehavior]TrafficControllerGenFunc
	mux   sync.Mutex
}

// NewSafeTrafficGenFuncs returns a new SafeTrafficGenFuncs instance.
func NewSafeTrafficGenFuncs() *SafeTrafficGenFuncs {
	return &SafeTrafficGenFuncs{
		funcs: make(map[ControlBehavior]TrafficControllerGenFunc),
		mux:   sync.Mutex{},
	}
}

// Store sets the value for a key.
func (s *SafeTrafficGenFuncs) Store(key ControlBehavior, value TrafficControllerGenFunc) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.funcs[key] = value
}

// Delete deletes the value for a key
func (s *SafeTrafficGenFuncs) Delete(key ControlBehavior) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.funcs, key)
}

// Load returns the value stored in the map for a key
func (s *SafeTrafficGenFuncs) Load(key ControlBehavior) (TrafficControllerGenFunc, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	v, ok := s.funcs[key]
	return v, ok
}

var (
	tcGenFuncMap = NewSafeTrafficGenFuncs()
	tcMap        = NewSafeTrafficControllers()
	tcMux        = new(sync.RWMutex)
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap.Store(Reject, func(rule *FlowRule) *TrafficShapingController {
		return NewTrafficShapingController(
			NewDefaultTrafficShapingCalculator(rule.Count),
			NewDefaultTrafficShapingChecker(rule.MetricType), rule,
		)
	})

	tcGenFuncMap.Store(Throttling, func(rule *FlowRule) *TrafficShapingController {
		return NewTrafficShapingController(
			NewDefaultTrafficShapingCalculator(rule.Count),
			NewThrottlingChecker(rule.MaxQueueingTimeMs), rule,
		)
	})
}

func logRuleUpdate(m *SafeTrafficControllers) {
	bs, err := json.Marshal(m.GetRules())
	if err != nil {
		logger.Warnf("[FlowRuleManager] Flow rules loaded error: %+v", err)
	} else {
		logger.Infof("[FlowRuleManager] Flow rules loaded: %s", bs)
	}
}

func onRuleUpdate(rules []*FlowRule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	m := buildFlowMap(rules)
	tcMap.Reset(m.controller)
	logRuleUpdate(m)

	return nil
}

// LoadRules loads the given flow rules to the rule manager, while all previous rules will be replaced.
func LoadRules(rules []*FlowRule) (bool, error) {
	// TODO: rethink the design
	err := onRuleUpdate(rules)
	return true, err
}

func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

func GetRules() []*FlowRule {
	return tcMap.GetRules()
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
	tcGenFuncMap.Store(cb, generator)
	return nil
}

func RemoveTrafficShapingGenerator(cb ControlBehavior) error {
	if cb >= Reject && cb <= WarmUpThrottling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcGenFuncMap.Delete(cb)
	return nil
}

func getTrafficControllerListFor(name string) []*TrafficShapingController {
	v, _ := tcMap.Load(name)
	return v
}

func buildFlowMap(rules []*FlowRule) *SafeTrafficControllers {
	m := NewSafeTrafficControllers()
	if len(rules) == 0 {
		return m
	}

	for _, rule := range rules {
		if err := IsValidFlowRule(rule); err != nil {
			logger.Warnf("Ignoring invalid flow rule: %v, reason: %s", rule, err.Error())
			continue
		}
		if rule.LimitOrigin == "" {
			rule.LimitOrigin = LimitOriginDefault
		}
		generator, supported := tcGenFuncMap.Load(rule.ControlBehavior)
		if !supported {
			logger.Warnf("Ignoring the rule due to unsupported control behavior: %v", rule)
			continue
		}
		tsc := generator(rule)
		if tsc == nil {
			logger.Warnf("Ignoring the rule due to bad generated traffic controller: %v", rule)
			continue
		}

		rulesOfRes, exists := m.Load(rule.Resource)
		if !exists {
			m.Store(rule.Resource, []*TrafficShapingController{tsc})
		} else {
			m.Store(rule.Resource, append(rulesOfRes, tsc))
		}
	}
	return m
}

// IsValidFlowRule checks whether the given FlowRule is valid.
func IsValidFlowRule(rule *FlowRule) error {
	if rule == nil {
		return errors.New("nil FlowRule")
	}
	if rule.Resource == "" {
		return errors.New("empty resource name")
	}
	if rule.Count < 0 {
		return errors.New("negative threshold")
	}
	if rule.MetricType < 0 {
		return errors.New("invalid metric type")
	}
	if rule.RelationStrategy < 0 {
		return errors.New("invalid relation strategy")
	}
	if rule.ControlBehavior < 0 {
		return errors.New("invalid control behavior")
	}

	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return errors.New("Bad flow rule: invalid control behavior")
	}
	if rule.ClusterMode && rule.ID <= 0 {
		return errors.New("invalid cluster rule ID")
	}

	return checkControlBehaviorField(rule)
}

func checkControlBehaviorField(rule *FlowRule) error {
	switch rule.ControlBehavior {
	case WarmUp:
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("invalid warmUpPeriodSec")
		}
	case WarmUpThrottling:
		if rule.WarmUpPeriodSec <= 0 {
			return errors.New("invalid warmUpPeriodSec")
		}
	}
	return nil
}
