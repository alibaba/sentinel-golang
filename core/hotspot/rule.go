package hotspot

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/alibaba/sentinel-golang/logging"
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
	KindInt ParamKind = iota
	KindString
	KindBool
	KindFloat64
	KindSum
)

func (t ParamKind) String() string {
	switch t {
	case KindInt:
		return "KindInt"
	case KindString:
		return "KindString"
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
	ValKind   ParamKind `json:"valKind"`
	ValStr    string    `json:"valStr"`
	Threshold int64     `json:"threshold"`
}

func (s *SpecificValue) String() string {
	return fmt.Sprintf("SpecificValue: [ValKind: %+v, ValStr: %s]", s.ValKind, s.ValStr)
}

// Rule represents the hotspot(frequent) parameter flow control rule
type Rule struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// Resource is the resource name
	Resource        string          `json:"resource"`
	MetricType      MetricType      `json:"metricType"`
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	// ParamIndex is the index in context arguments slice.
	ParamIndex int     `json:"paramIndex"`
	Threshold  float64 `json:"threshold"`
	// MaxQueueingTimeMs only take effect in both Throttling ControlBehavior and QPS MetricType
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount is the silent count
	// Only take effect in both Reject ControlBehavior and QPS MetricType
	BurstCount int64 `json:"burstCount"`
	// DurationInSec is the time interval in statistic
	// Only take effect in QPS MetricType
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity is the max capacity of cache statistic
	ParamsMaxCapacity int64 `json:"paramsMaxCapacity"`
	// SpecificItems indicates the special threshold for specific value
	SpecificItems []SpecificValue `json:"specificItems"`
}

func (r *Rule) String() string {
	return fmt.Sprintf("{Id:%s, Resource:%s, MetricType:%+v, ControlBehavior:%+v, ParamIndex:%d, Threshold:%f, MaxQueueingTimeMs:%d, BurstCount:%d, DurationInSec:%d, ParamsMaxCapacity:%d, SpecificItems:%+v}",
		r.ID, r.Resource, r.MetricType, r.ControlBehavior, r.ParamIndex, r.Threshold, r.MaxQueueingTimeMs, r.BurstCount, r.DurationInSec, r.ParamsMaxCapacity, r.SpecificItems)
}
func (r *Rule) ResourceName() string {
	return r.Resource
}

// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
func (r *Rule) IsStatReusable(newRule *Rule) bool {
	return r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.DurationInSec == newRule.DurationInSec
}

// Equals checks whether current rule is consistent with the given rule.
func (r *Rule) Equals(newRule *Rule) bool {
	baseCheck := r.Resource == newRule.Resource && r.MetricType == newRule.MetricType && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.ParamIndex == newRule.ParamIndex && r.Threshold == newRule.Threshold && r.DurationInSec == newRule.DurationInSec && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems)
	if !baseCheck {
		return false
	}
	if r.ControlBehavior == Reject {
		return r.BurstCount == newRule.BurstCount
	} else if r.ControlBehavior == Throttling {
		return r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs
	} else {
		return false
	}
}

// parseSpecificItems parses the SpecificValue as real value.
func parseSpecificItems(source []SpecificValue) map[interface{}]int64 {
	ret := make(map[interface{}]int64)
	if len(source) == 0 {
		return ret
	}
	for _, item := range source {
		switch item.ValKind {
		case KindInt:
			realVal, err := strconv.Atoi(item.ValStr)
			if err != nil {
				logging.Errorf("Failed to parse value for int specific item. paramKind: %+v, value: %s, err: %+v", item.ValKind, item.ValStr, err)
				continue
			}
			ret[realVal] = item.Threshold

		case KindString:
			ret[item.ValStr] = item.Threshold

		case KindBool:
			realVal, err := strconv.ParseBool(item.ValStr)
			if err != nil {
				logging.Errorf("Failed to parse value for bool specific item. value: %s, err: %+v", item.ValStr, err)
				continue
			}
			ret[realVal] = item.Threshold

		case KindFloat64:
			realVal, err := strconv.ParseFloat(item.ValStr, 64)
			if err != nil {
				logging.Errorf("Failed to parse value for float specific item. value: %s, err: %+v", item.ValStr, err)
				continue
			}
			realVal, err = strconv.ParseFloat(fmt.Sprintf("%.5f", realVal), 64)
			if err != nil {
				logging.Errorf("Failed to parse value for float specific item. value: %s, err: %+v", item.ValStr, err)
				continue
			}
			ret[realVal] = item.Threshold
		default:
			logging.Errorf("Unsupported kind for specific item: %d", item.ValKind)
		}
	}
	return ret
}
