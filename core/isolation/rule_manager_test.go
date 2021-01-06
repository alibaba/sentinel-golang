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

func clearData() {
	ruleMap = make(map[string][]*Rule)
	currentRules = make(map[string][]*Rule, 0)
}

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

		clearData()
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
		clearData()
	})

	t.Run("TestClearRules_normal", func(t *testing.T) {
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

		succ, err := LoadRules([]*Rule{r1, r2, r3})
		assert.True(t, succ && err == nil)

		assert.True(t, ClearRules() == nil)

		assert.True(t, len(ruleMap["abc1"]) == 0)
		assert.True(t, len(currentRules["abc1"]) == 0)
		assert.True(t, len(ruleMap["abc2"]) == 0)
		assert.True(t, len(currentRules["abc2"]) == 0)
		assert.True(t, len(ruleMap["abc3"]) == 0)
		assert.True(t, len(currentRules["abc3"]) == 0)
		clearData()
	})
}

func TestLoadRulesOfResource(t *testing.T) {
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

	r4 := &Rule{
		Resource:   "abc2",
		MetricType: MetricType(1),
		Threshold:  300,
	}

	succ, err := LoadRules([]*Rule{r1, r2, r3, r4})
	assert.True(t, succ && err == nil)

	t.Run("LoadRulesOfResource_empty_resource", func(t *testing.T) {
		succ, err = LoadRulesOfResource("", []*Rule{r1, r2})
		assert.True(t, !succ && err != nil)
	})

	t.Run("LoadRulesOfResource_cache_hit", func(t *testing.T) {
		r111 := *r2
		r122 := *r4
		succ, err = LoadRulesOfResource("abc2", []*Rule{&r111, &r122})
		assert.True(t, !succ && err == nil)
	})

	t.Run("LoadRulesOfResource_clear", func(t *testing.T) {
		succ, err = LoadRulesOfResource("abc1", []*Rule{})
		assert.True(t, succ && err == nil)
		assert.True(t, len(ruleMap["abc1"]) == 0 && len(currentRules["abc1"]) == 0)
		assert.True(t, len(ruleMap["abc2"]) == 1 && len(currentRules["abc2"]) == 2)
	})
	clearData()
}

func Test_ResourceRuleUpdate(t *testing.T) {
	logging.ResetGlobalLoggerLevel(logging.DebugLevel)
	t.Run("Test_onResourceRuleUpdate_normal", func(t *testing.T) {
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

		succ, err := LoadRules([]*Rule{r1, r2, r3})
		assert.True(t, succ && err == nil)

		r111 := r1
		r111.Threshold = 100
		err = onResourceRuleUpdate("abc1", []*Rule{r111})

		assert.True(t, err == nil)
		assert.True(t, len(ruleMap["abc1"]) == 1)
		assert.True(t, len(currentRules["abc1"]) == 1)
		assert.True(t, ruleMap["abc1"][0] == r111)

		assert.True(t, len(ruleMap["abc2"]) == 1)
		assert.True(t, len(currentRules["abc2"]) == 1)

		clearData()
	})

	t.Run("TestClearRulesOfResource_normal", func(t *testing.T) {
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

		succ, err := LoadRules([]*Rule{r1, r2, r3})
		assert.True(t, succ && err == nil)

		assert.True(t, ClearRulesOfResource("abc1") == nil)

		assert.True(t, len(ruleMap["abc1"]) == 0)
		assert.True(t, len(currentRules["abc1"]) == 0)
		assert.True(t, len(ruleMap["abc2"]) == 1)
		assert.True(t, len(currentRules["abc2"]) == 1)
		assert.True(t, len(ruleMap["abc3"]) == 0)
		assert.True(t, len(currentRules["abc3"]) == 1)
		clearData()
	})
}
