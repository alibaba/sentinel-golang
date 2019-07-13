package flow

import (
	"context"
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core/node"
)

const (
	LimitAppDefault     = "default"
	LimitAppOther       = "other"
	ResourceNameDefault = "default"
)

type StrategyType int8

const (
	StrategyDirect StrategyType = iota
	StrategyRelate
	StrategyChain
)

type ControlBehaviorType int8

const (
	ControlBehaviorDefault ControlBehaviorType = iota
	ControlBehaviorWarmUp
	ControlBehaviorRateLimiter
	ControlBehaviorWarmUpRateLimiter
)

type FlowGradeType int8

const (
	FlowGradeThread FlowGradeType = iota
	FlowGradeQps
)

type RuleChecker interface {
	passCheck(ctx context.Context, node *node.Node, count int) bool
}

type rule struct {
	resource_ string
	limitApp_ string
	grade_    FlowGradeType
	//  Flow control threshold count.
	count_           uint64
	strategy_        StrategyType
	refResource_     string
	controlBehavior_ ControlBehaviorType
	warmUpPeriodSec_ int32
	/**
	 * Max queueing time in rate limiter behavior.
	 */
	maxQueueingTimeMs_ int32
	controller_        TrafficShapingController
}

func (r *rule) PassCheck(ctx context.Context, node node.Node, count int) bool {

	return true
}

func (r *rule) validate() error {
	return nil
}

func newRuleBuilder() *rule {
	return &rule{
		resource_:          ResourceNameDefault,
		limitApp_:          LimitAppDefault,
		grade_:             FlowGradeQps,
		count_:             100,
		strategy_:          StrategyDirect,
		refResource_:       "",
		controlBehavior_:   ControlBehaviorRateLimiter,
		warmUpPeriodSec_:   10,
		maxQueueingTimeMs_: 500,
	}
}

func (r *rule) resource(resource string) *rule {
	r.resource_ = resource
	return r
}
func (r *rule) limitApp(limitApp string) *rule {
	r.limitApp_ = limitApp
	return r
}
func (r *rule) grade(grade_ FlowGradeType) *rule {
	r.grade_ = grade_
	return r
}
func (r *rule) count(count uint64) *rule {
	r.count_ = count
	return r
}
func (r *rule) strategy(strategy StrategyType) *rule {
	r.strategy_ = strategy
	return r
}
func (r *rule) refResource(refResource string) *rule {
	r.refResource_ = refResource
	return r
}
func (r *rule) controlBehavior(controlBehavior ControlBehaviorType) *rule {
	r.controlBehavior_ = controlBehavior
	return r
}
func (r *rule) warmUpPeriodSec(warmUpPeriodSec int32) *rule {
	r.warmUpPeriodSec_ = warmUpPeriodSec
	return r
}
func (r *rule) maxQueueingTimeMs(maxQueueingTimeMs int32) *rule {
	r.maxQueueingTimeMs_ = maxQueueingTimeMs
	return r
}
func (r *rule) controller(controller TrafficShapingController) *rule {
	r.controller_ = controller
	return r
}

type RuleManager struct {
	flowRules map[string][]*rule
}

func NewRuleManager() *RuleManager {
	return &RuleManager{
		flowRules: nil,
	}
}

func LoadRules(rm *RuleManager, rules []*rule) {
	if len(rules) == 0 {
		println("RuleManager| load empty rule ")
		return
	}
	rm.flowRules = buildFlowRuleMap(rules)
}

// Get a copy of the rules.
func (rm *RuleManager) getAllRule() []*rule {

	ret := make([]*rule, 0)
	for _, value := range rm.flowRules {
		ret = append(ret, value...)
	}
	return ret
}

func (rm *RuleManager) getRuleMap() map[string][]*rule {
	return rm.flowRules
}

func (rm *RuleManager) getRuleBySource(resource string) []*rule {
	rules := rm.flowRules[resource]
	if len(rules) == 0 {
		return nil
	}
	return rules
}

func buildFlowRuleMap(rules []*rule) map[string][]*rule {
	ret := make(map[string][]*rule)

	for _, r := range rules {
		err := r.validate()
		if err != nil {
			fmt.Printf("validate rule  fail, the reason is %s ", err.Error())
			continue
		}
		r.controller_ = generateFlowControl(r)
		srcName := r.resource_
		var slc = ret[srcName]
		if slc == nil {
			slc = make([]*rule, 0)
		}
		slc = append(slc, r)
		ret[srcName] = slc
	}
	return ret
}

func generateFlowControl(r *rule) TrafficShapingController {
	if r.grade_ == FlowGradeQps {
		switch r.controlBehavior_ {
		case ControlBehaviorWarmUp:
			return new(WarmUpController)
		case ControlBehaviorRateLimiter:
			return new(RateLimiterController)
		case ControlBehaviorWarmUpRateLimiter:
			return new(WarmUpRateLimiterController)
		default:
		}
	}
	return new(DefaultController)
}

//
//type FlowPropertyListener struct {
//}
//
//func (fpl *FlowPropertyListener) ConfigUpdate(value interface{}) {
//
//}
//func (fpl *FlowPropertyListener) ConfigLoad(value interface{}) {
//
//}
//
