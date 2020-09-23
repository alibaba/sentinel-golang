package flow

import (
	"encoding/json"
	"fmt"
)

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

// Rule describes the strategy of flow control, the flow control strategy is based on QPS statistic metric
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty"`
	// Resource represents the resource name.
	Resource               string                 `json:"resource"`
	TokenCalculateStrategy TokenCalculateStrategy `json:"tokenCalculateStrategy"`
	ControlBehavior        ControlBehavior        `json:"controlBehavior"`
	Threshold              float64                `json:"threshold"`
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
		return fmt.Sprintf("Rule{Resource=%s, TokenCalculateStrategy=%s, ControlBehavior=%s, "+
			"Threshold=%.2f, RelationStrategy=%s, RefResource=%s, MaxQueueingTimeMs=%d, WarmUpPeriodSec=%d, WarmUpColdFactor=%d}",
			r.Resource, r.TokenCalculateStrategy, r.ControlBehavior, r.Threshold, r.RelationStrategy,
			r.RefResource, r.MaxQueueingTimeMs, r.WarmUpPeriodSec, r.WarmUpColdFactor)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
