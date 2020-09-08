package flow

import (
	"encoding/json"
	"fmt"
)

// MetricType represents the target metric type.
type MetricType int32

const (
	// Concurrency represents concurrency count.
	Concurrency MetricType = iota
	// QPS represents request count per second.
	QPS
)

func (s MetricType) String() string {
	switch s {
	case Concurrency:
		return "Concurrency"
	case QPS:
		return "QPS"
	default:
		return "Undefined"
	}
}

// RelationStrategy indicates the flow control strategy based on the relation of invocations.
type RelationStrategy int32

const (
	// CurrentResource means flow control by current resource directly.
	CurrentResource RelationStrategy = iota
	// AssociatedResource means flow control by the associated resource rather than current resource.
	AssociatedResource
)

func (s RelationStrategy) String() string {
	switch s {
	case CurrentResource:
		return "CurrentResource"
	case AssociatedResource:
		return "AssociatedResource"
	default:
		return "Undefined"
	}
}

type TokenCalculateStrategy int32

const (
	Direct TokenCalculateStrategy = iota
	WarmUp
)

func (s TokenCalculateStrategy) String() string {
	switch s {
	case Direct:
		return "Direct"
	case WarmUp:
		return "WarmUp"
	default:
		return "Undefined"
	}
}

type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	Throttling
)

func (s ControlBehavior) String() string {
	switch s {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return "Undefined"
	}
}

// Rule describes the strategy of flow control.
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID uint64 `json:"id,omitempty"`

	// Resource represents the resource name.
	Resource               string                 `json:"resource"`
	MetricType             MetricType             `json:"metricType"`
	TokenCalculateStrategy TokenCalculateStrategy `json:"tokenCalculateStrategy"`
	ControlBehavior        ControlBehavior        `json:"controlBehavior"`
	Count                  float64                `json:"count"`
	RelationStrategy       RelationStrategy       `json:"relationStrategy"`
	RefResource            string                 `json:"refResource"`
	MaxQueueingTimeMs      uint32                 `json:"maxQueueingTimeMs"`
	WarmUpPeriodSec        uint32                 `json:"warmUpPeriodSec"`
	WarmUpColdFactor       uint32                 `json:"warmUpColdFactor"`
}

func (r *Rule) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Rule{Resource=%s, MetricType=%s, TokenCalculateStrategy=%s, ControlBehavior=%s, "+
			"Count=%.2f, RelationStrategy=%s, WarmUpPeriodSec=%d, WarmUpColdFactor=%d, MaxQueueingTimeMs=%d}",
			r.Resource, r.MetricType, r.TokenCalculateStrategy, r.ControlBehavior, r.Count, r.RelationStrategy, r.WarmUpPeriodSec, r.WarmUpColdFactor, r.MaxQueueingTimeMs)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
