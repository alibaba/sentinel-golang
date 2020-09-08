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

// RelationStrategy indicates the flow control strategy based on the relation of invocations.
type RelationStrategy int32

const (
	// Direct means flow control by current resource directly.
	Direct RelationStrategy = iota
	// AssociatedResource means flow control by the associated resource rather than current resource.
	AssociatedResource
)

// ControlBehavior indicates the traffic shaping behaviour.
type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	WarmUp
	Throttling
	WarmUpThrottling
)

// Rule describes the strategy of flow control.
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID uint64 `json:"id,omitempty"`

	// Resource represents the resource name.
	Resource   string     `json:"resource"`
	MetricType MetricType `json:"metricType"`
	// Count represents the threshold.
	Count           float64         `json:"count"`
	ControlBehavior ControlBehavior `json:"controlBehavior"`

	RelationStrategy  RelationStrategy `json:"relationStrategy"`
	RefResource       string           `json:"refResource"`
	MaxQueueingTimeMs uint32           `json:"maxQueueingTimeMs"`
	WarmUpPeriodSec   uint32           `json:"warmUpPeriodSec"`
	WarmUpColdFactor  uint32           `json:"warmUpColdFactor"`
}

func (f *Rule) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Rule{resource=%s, id=%d, metricType=%d, threshold=%.2f}",
			f.Resource, f.ID, f.MetricType, f.Count)
	}
	return string(b)
}

func (f *Rule) ResourceName() string {
	return f.Resource
}
