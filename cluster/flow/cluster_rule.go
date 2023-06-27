package flow

import (
	"encoding/json"
	"fmt"
)

type ClusterRule struct {
	// ID represents the unique ID of the rule (optional).
	// @Optional
	ID string `json:"id,omitempty"`
	// Resource represents the resource name.
	// @Required
	Resource string `json:"resource"`
	// Threshold means the threshold during StatIntervalInMs.
	// If StatIntervalInMs is 1000(1 second), Threshold means QPS.
	// @Required
	Threshold int64 `json:"threshold"`
	// StatIntervalInMs indicates the statistic interval and it's the required setting for flow ClusterRule.
	// @Required
	StatIntervalInMs uint32 `json:"statIntervalInMs"`
	// ClusterTokenSequence indicates the number of tokens fetched from the Token Server each time.
	// If TokenSequence <=1 means.
	// @Optional
	TokenSequence uint32 `json:"tokenSequence"`
}

func (r *ClusterRule) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("ClusterRule{ID=%s, Resource=%s, Threshold=%d, StatIntervalInMs=%d, TokenSequence=%d}",
			r.ID, r.Resource, r.Threshold, r.StatIntervalInMs, r.TokenSequence)
	}
	return string(b)
}

func (r *ClusterRule) ResourceName() string {
	return r.Resource
}
