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

package datasource

import (
	"fmt"
	"strconv"

	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

type HotspotRule struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// Resource is the resource name
	Resource string `json:"resource"`
	// MetricType indicates the metric type for checking logic.
	// For Concurrency metric, hotspot module will check the each hot parameter's concurrency,
	//		if concurrency exceeds the Threshold, reject the traffic directly.
	// For QPS metric, hotspot module will check the each hot parameter's QPS,
	//		the ControlBehavior decides the behavior of traffic shaping controller
	MetricType hotspot.MetricType `json:"metricType"`
	// ControlBehavior indicates the traffic shaping behaviour.
	// ControlBehavior only takes effect when MetricType is QPS
	ControlBehavior hotspot.ControlBehavior `json:"controlBehavior"`
	// ParamIndex is the index in context arguments slice.
	// if ParamIndex is great than or equals to zero, ParamIndex means the <ParamIndex>-th parameter
	// if ParamIndex is the negative, ParamIndex means the reversed <ParamIndex>-th parameter
	ParamIndex int `json:"paramIndex"`
	// Threshold is the threshold to trigger rejection
	Threshold int64 `json:"threshold"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling and MetricType is QPS
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount is the silent count
	// BurstCount only takes effect when ControlBehavior is Reject and MetricType is QPS
	BurstCount int64 `json:"burstCount"`
	// DurationInSec is the time interval in statistic
	// DurationInSec only takes effect when MetricType is QPS
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity is the max capacity of cache statistic
	ParamsMaxCapacity int64           `json:"paramsMaxCapacity"`
	SpecificItems     []SpecificValue `json:"specificItems"`
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

// arseSpecificItems parses the SpecificValue as real value.
func parseSpecificItems(source []SpecificValue) map[interface{}]int64 {
	ret := make(map[interface{}]int64, len(source))
	if len(source) == 0 {
		return ret
	}
	for _, item := range source {
		switch item.ValKind {
		case KindInt:
			realVal, err := strconv.Atoi(item.ValStr)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ValKind, "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold

		case KindString:
			ret[item.ValStr] = item.Threshold

		case KindBool:
			realVal, err := strconv.ParseBool(item.ValStr)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for bool specific item", "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold

		case KindFloat64:
			realVal, err := strconv.ParseFloat(item.ValStr, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValStr", item.ValStr)
				continue
			}
			realVal, err = strconv.ParseFloat(fmt.Sprintf("%.5f", realVal), 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold
		default:
			logging.Error(errors.New("Unsupported kind for specific item"), "", item.ValKind)
		}
	}
	return ret
}
