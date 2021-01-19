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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricType_String(t *testing.T) {
	t.Run("TestMetricType_String", func(t *testing.T) {
		assert.True(t, fmt.Sprintf("%+v", Concurrency) == "Concurrency")

	})
}

func Test_Rule_String(t *testing.T) {
	t.Run("Test_Rule_String_Normal", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific["sss"] = 1
		specific["1123"] = 3
		r := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			ParamKey:          "key",
			Threshold:         110.0,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
			SpecificItems:     specific,
		}
		assert.True(t, fmt.Sprintf("%+v", []*Rule{r}) == "[{Id:abc, Resource:abc, MetricType:Concurrency, ControlBehavior:Reject, ParamIndex:0, ParamKey:key, Threshold:110, MaxQueueingTimeMs:5, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:map[1123:3 sss:1]}]")
	})
}

func Test_Rule_Equals(t *testing.T) {
	t.Run("Test_Rule_Equals", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific["sss"] = 1
		specific[1123] = 3
		r1 := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			ParamKey:          "testKey",
			Threshold:         110.0,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
			SpecificItems:     specific,
		}

		specific2 := make(map[interface{}]int64)
		specific2["sss"] = 1
		specific2[1123] = 3
		r2 := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			ParamKey:          "testKey",
			Threshold:         110.0,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
			SpecificItems:     specific2,
		}
		assert.True(t, r1.Equals(r2))
	})
}
