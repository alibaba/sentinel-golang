// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hotspot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// ControlBehavior indicates the traffic shaping behaviour.
type ControlBehavior int32

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
type MetricType int32

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

// Rule represents the hotspot(frequent) parameter flow control rule
type Rule struct {
	// ID is the unique id
	ID string `json:"Id,omitempty"`
	// Resource is the resource name
	Resource string `json:"Resource"`
	// MetricType indicates the metric type for checking logic.
	// For Concurrency metric, hotspot module will check the each hot parameter's concurrency,
	//		if concurrency exceeds the Threshold, reject the traffic directly.
	// For QPS metric, hotspot module will check the each hot parameter's QPS,
	//		the ControlBehavior decides the behavior of traffic shaping controller
	MetricType MetricType `json:"MetricType"`
	// ControlBehavior indicates the traffic shaping behaviour.
	// ControlBehavior only takes effect when MetricType is QPS
	ControlBehavior ControlBehavior `json:"ControlBehavior"`
	// ParamIndex is the index in context arguments slice.
	// if ParamIndex is great than or equals to zero, ParamIndex means the <ParamIndex>-th parameter
	// if ParamIndex is the negative, ParamIndex means the reversed <ParamIndex>-th parameter
	ParamIndex int `json:"ParamIndex"`
	// ParamKey is the key in EntryContext.Input.Attachments map.
	// ParamKey can be used as a supplement to ParamIndex to facilitate rules to quickly obtain parameter from a large number of parameters
	// ParamKey is mutually exclusive with ParamIndex, ParamKey has the higher priority than ParamIndex
	ParamKey string `json:"ParamKey"`
	// Threshold is the threshold to trigger rejection
	Threshold int64 `json:"Threshold"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling and MetricType is QPS
	MaxQueueingTimeMs int64 `json:"MaxQueueingTimeMs"`
	// BurstCount is the silent count
	// BurstCount only takes effect when ControlBehavior is Reject and MetricType is QPS
	BurstCount int64 `json:"BurstCount"`
	// DurationInSec is the time interval in statistic
	// DurationInSec only takes effect when MetricType is QPS
	DurationInSec int64 `json:"DurationInSec"`
	// ParamsMaxCapacity is the max capacity of cache statistic
	ParamsMaxCapacity int64 `json:"ParamsMaxCapacity"`
	// SpecificItems indicates the special threshold for specific value
	SpecificItems map[interface{}]int64
}

type SpecificItem struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

func (t MetricType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
func (c ControlBehavior) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// Store the key-value pairs in specificItems in a temporary slice
func convertSpecificItems(specificItems map[interface{}]int64) []*SpecificItem {
	items := make([]*SpecificItem, 0, len(specificItems))
	for k, v := range specificItems {
		key, ok := k.(string)
		if !ok {
			key = fmt.Sprintf("%v", k)
		}
		item := &SpecificItem{
			Key:   key,
			Value: v,
		}
		items = append(items, item)
	}
	return items
}

func (r *Rule) String() string {
	type tempRule Rule
	specificItems := convertSpecificItems(r.SpecificItems)

	b, err := json.Marshal(&struct {
		*tempRule
		SpecificItems []*SpecificItem `json:"SpecificItems"`
	}{
		tempRule:      (*tempRule)(r),
		SpecificItems: specificItems,
	})
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("{Id:%s, Resource:%s, MetricType:%+v, ControlBehavior:%+v, ParamIndex:%d, ParamKey:%s, Threshold:%d, MaxQueueingTimeMs:%d, BurstCount:%d, DurationInSec:%d, ParamsMaxCapacity:%d, SpecificItems:%+v}",
			r.ID, r.Resource, r.MetricType, r.ControlBehavior, r.ParamIndex, r.ParamKey, r.Threshold, r.MaxQueueingTimeMs, r.BurstCount, r.DurationInSec, r.ParamsMaxCapacity, r.SpecificItems)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}

// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
func (r *Rule) IsStatReusable(newRule *Rule) bool {
	return r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.DurationInSec == newRule.DurationInSec && r.MetricType == newRule.MetricType
}

// Equals checks whether current rule is consistent with the given rule.
func (r *Rule) Equals(newRule *Rule) bool {
	baseCheck := r.Resource == newRule.Resource && r.MetricType == newRule.MetricType && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.ParamIndex == newRule.ParamIndex && r.ParamKey == newRule.ParamKey && r.Threshold == newRule.Threshold && r.DurationInSec == newRule.DurationInSec && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems)
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
