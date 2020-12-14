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

package isolation

import (
	"testing"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/stretchr/testify/assert"
)

func TestLoadRules(t *testing.T) {
	t.Run("TestLoadRules_1", func(t *testing.T) {
		logging.ResetGlobalLoggerLevel(logging.DebugLevel)
		r1 := &Rule{
			Resource:   "abc1",
			MetricType: Concurrency,
			Threshold:  100,
		}
		r2 := &Rule{
			Resource:   "abc2",
			MetricType: Concurrency,
			Threshold:  200,
		}
		r3 := &Rule{
			Resource:   "abc3",
			MetricType: MetricType(1),
			Threshold:  200,
		}
		_, err := LoadRules([]*Rule{r1, r2, r3})
		assert.True(t, err == nil)
		assert.True(t, len(ruleMap) == 2)
		assert.True(t, len(ruleMap["abc1"]) == 1)
		assert.True(t, ruleMap["abc1"][0] == r1)
		assert.True(t, len(ruleMap["abc2"]) == 1)
		assert.True(t, ruleMap["abc2"][0] == r2)

		err = ClearRules()
		assert.True(t, err == nil)
	})

	t.Run("loadSameRules", func(t *testing.T) {
		_, err := LoadRules([]*Rule{
			{
				Resource:   "abc1",
				MetricType: Concurrency,
				Threshold:  100,
			},
		})
		assert.Nil(t, err)
		ok, err := LoadRules([]*Rule{
			{
				Resource:   "abc1",
				MetricType: Concurrency,
				Threshold:  100,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}
