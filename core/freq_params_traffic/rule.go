package freq_params_traffic

import (
	"fmt"
	"reflect"
	"strconv"
)

// ControlBehavior indicates the traffic shaping behaviour.
type ControlBehavior int8

const (
	Reject ControlBehavior = iota
	Throttling
)

func (t ControlBehavior) String() string {
	switch t {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return strconv.Itoa(int(t))
	}
}

// MetricType represents the target metric type.
type MetricType int8

const (
	// Concurrency represents concurrency count.
	Concurrency MetricType = iota
	// QPS represents request count per second.
	QPS
)

func (t MetricType) String() string {
	switch t {
	case Concurrency:
		return "Concurrency"
	case QPS:
		return "QPS"
	default:
		return "Undefined"
	}
}

// ParamKind represents the Param kind.
type ParamKind int

const (
	kindInt ParamKind = iota
	kindString
	KindBool
	KindFloat64
	KindSum
)

func (t ParamKind) String() string {
	switch t {
	case kindInt:
		return "kindInt"
	case kindString:
		return "kindString"
	case KindBool:
		return "KindBool"
	case KindFloat64:
		return "KindFloat64"
	default:
		return "Undefined"
	}
}

// SpecificValue indicates the specific param, contain the supported param kind and concrete value.
type SpecificValue struct {
	ValKind ParamKind
	ValStr  string
}

func (s *SpecificValue) String() string {
	return fmt.Sprintf("SpecificValue:[ValKind: %+v, ValStr: %s]", s.ValKind, s.ValStr)
}

// Rule represents the frequency parameter flow control rule
type Rule struct {
	// Id is the unique id
	Id string
	// Resource is the resource name
	Resource   string
	MetricType MetricType
	Behavior   ControlBehavior
	// ParamIndex is the index in context arguments slice.
	ParamIndex int
	Threshold  float64
	// MaxQueueingTimeMs is the max queueing time in Throttling ControlBehavior
	MaxQueueingTimeMs int64
	BurstCount        int64
	DurationInSec     int64
	ParamsMaxCapacity int64
	SpecificItems     map[SpecificValue]int64
}

func (r *Rule) String() string {
	return fmt.Sprintf("{Id:%s, Resource:%s, MetricType:%+v, Behavior:%+v, ParamIndex:%d, Threshold:%f, MaxQueueingTimeMs:%d, BurstCount:%d, DurationInSec:%d, ParamsMaxCapacity:%d, SpecificItems:%+v},",
		r.Id, r.Resource, r.MetricType, r.Behavior, r.ParamIndex, r.Threshold, r.MaxQueueingTimeMs, r.BurstCount, r.DurationInSec, r.ParamsMaxCapacity, r.SpecificItems)
}
func (r *Rule) ResourceName() string {
	return r.Resource
}

// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
func (r *Rule) IsStatReusable(newRule *Rule) bool {
	return r.Resource == newRule.Resource && r.Behavior == newRule.Behavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.DurationInSec == newRule.DurationInSec
}

// IsEqualsTo checks whether current rule is consistent with the given rule.
func (r *Rule) Equals(newRule *Rule) bool {
	baseCheck := r.Resource == newRule.Resource && r.MetricType == newRule.MetricType && r.Behavior == newRule.Behavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.ParamIndex == newRule.ParamIndex && r.Threshold == newRule.Threshold && r.DurationInSec == newRule.DurationInSec && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems)
	if !baseCheck {
		return false
	}
	if r.Behavior == Reject {
		return r.BurstCount == newRule.BurstCount
	} else if r.Behavior == Throttling {
		return r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs
	} else {
		return false
	}
}

// parseSpecificItems parses the SpecificValue as real value.
func parseSpecificItems(source map[SpecificValue]int64) map[interface{}]int64 {
	ret := make(map[interface{}]int64)
	if len(source) == 0 {
		return ret
	}
	for k, v := range source {
		switch k.ValKind {
		case kindInt:
			realVal, err := strconv.Atoi(k.ValStr)
			if err != nil {
				logger.Errorf("Fail to parse value for int specific item. paramKind: %+v, value: %s, err: %+v", k.ValKind, k.ValStr, err)
				continue
			}
			ret[realVal] = v

		case kindString:
			ret[k.ValStr] = v

		case KindBool:
			realVal, err := strconv.ParseBool(k.ValStr)
			if err != nil {
				logger.Errorf("Fail to parse value for int specific item. value: %s, err: %+v", k.ValStr, err)
				continue
			}
			ret[realVal] = v

		case KindFloat64:
			realVal, err := strconv.ParseFloat(k.ValStr, 64)
			if err != nil {
				logger.Errorf("Fail to parse value for int specific item. value: %s, err: %+v", k.ValStr, err)
				continue
			}
			realVal, err = strconv.ParseFloat(fmt.Sprintf("%.5f", realVal), 64)
			if err != nil {
				logger.Errorf("Fail to parse value for int specific item. value: %s, err: %+v", k.ValStr, err)
				continue
			}
			ret[realVal] = v

		default:
			logger.Errorf("Unsupported kind(%d) for specific item.", k.ValKind)
		}
	}
	return ret
}
