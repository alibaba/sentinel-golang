// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package outlier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

func clearData() {
	// resource name ---> outlier ejection rule
	outlierRules = make(map[string]*Rule)
	// resource name ---> circuitbreaker rule
	breakerRules = make(map[string]*circuitbreaker.Rule)
	// resource name ---> address ---> circuitbreaker
	nodeBreakers = make(map[string]map[string]circuitbreaker.CircuitBreaker)
	// resource name ---> outlier ejection rule
	currentRules = make(map[string]*Rule)
}

func Test_onRuleUpdateInvalid(t *testing.T) {
	r1 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.5, // MaxEjectionPercent should be in the range [0.0, 1.0]
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	resRulesMap := make(map[string]*Rule)
	resRulesMap[r1.Resource] = r1
	err := onRuleUpdate(resRulesMap)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(GetRules()))
	clearData()
}

func TestGetRules(t *testing.T) {
	r1 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	_, _ = LoadRules([]*Rule{r1})
	rules := GetRules()
	assert.True(t, len(rules) == 1 && rules[0].Resource == r1.Resource && rules[0].Strategy == r1.Strategy)
	clearData()
}

func TestGetNodeBreakersOfResource(t *testing.T) {
	r1 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	_, _ = LoadRules([]*Rule{r1})
	addNodeBreakerOfResource(r1.Resource, "node0")
	cbs := getNodeBreakersOfResource(r1.Resource)
	assert.True(t, len(cbs) == 1 && cbs["node0"].BoundRule() == r1.Rule)
	clearData()
}

func TestLoadRules(t *testing.T) {
	r1 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	_, err := LoadRules([]*Rule{r1})
	assert.Nil(t, err)
	ok, err := LoadRules([]*Rule{r1})
	assert.Nil(t, err)
	assert.False(t, ok)
	clearData()
}

func getTestRules() []*Rule {
	r1 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.SlowRequestRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	r2 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "example.helloworld",
			Strategy:         circuitbreaker.ErrorRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        1.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	r3 := &Rule{
		Rule: &circuitbreaker.Rule{
			Resource:         "test.resource",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 1,
			StatIntervalMs:   1000,
			Threshold:        10.0,
		},
		EnableActiveRecovery: true,
		MaxEjectionPercent:   1.0,
		RecoveryIntervalMs:   2000,
		MaxRecoveryAttempts:  5,
	}
	return []*Rule{r1, r2, r3}
}

func TestLoadRuleOfResource(t *testing.T) {
	rules := getTestRules()
	r1, r2, _ := rules[0], rules[1], rules[2]
	succ, err := LoadRules(rules)
	assert.Equal(t, 2, len(breakerRules))
	assert.True(t, succ && err == nil)

	t.Run("LoadRuleOfResource_empty_resource", func(t *testing.T) {
		succ, err = LoadRuleOfResource("", r1)
		assert.True(t, !succ && err != nil)
	})

	t.Run("LoadRuleOfResource_cache_hit", func(t *testing.T) {
		assert.Equal(t, r2, getOutlierRuleOfResource("example.helloworld"))
		succ, err = LoadRuleOfResource("example.helloworld", r1)
		assert.True(t, succ && err == nil)
	})

	t.Run("LoadRuleOfResource_clear", func(t *testing.T) {
		succ, err = LoadRuleOfResource("example.helloworld", nil)
		assert.Equal(t, 1, len(breakerRules))
		assert.True(t, succ && err == nil)
		assert.True(t, breakerRules["example.helloworld"] == nil && currentRules["example.helloworld"] == nil)
		assert.True(t, breakerRules["test.resource"] != nil && currentRules["test.resource"] != nil)
	})
	clearData()
}

func Test_onResourceRuleUpdate(t *testing.T) {
	rules := getTestRules()
	r1 := rules[0]
	succ, err := LoadRules(rules)
	addNodeBreakerOfResource(r1.Resource, "node0")
	assert.True(t, succ && err == nil)

	t.Run("Test_onResourceRuleUpdate_normal", func(t *testing.T) {
		r11 := r1
		r11.Threshold = 0.5
		assert.Nil(t, onResourceRuleUpdate(r1.Resource, r11))
		assert.Equal(t, getOutlierRuleOfResource(r1.Resource), r11)
		assert.Equal(t, 1, len(nodeBreakers[r1.Resource]))
		breakers := getNodeBreakersOfResource(r1.Resource)
		assert.Equal(t, breakers["node0"].BoundRule(), r11.Rule)
		clearData()
	})
}

func TestClearRuleOfResource(t *testing.T) {
	rules := getTestRules()
	r1 := rules[0]
	succ, err := LoadRules(rules)
	addNodeBreakerOfResource(r1.Resource, "node0")
	assert.True(t, succ && err == nil)

	t.Run("TestClearRuleOfResource_normal", func(t *testing.T) {
		assert.Equal(t, 1, len(nodeBreakers[r1.Resource]))
		assert.Nil(t, ClearRuleOfResource(r1.Resource))
		assert.Equal(t, 1, len(breakerRules))
		assert.Equal(t, 0, len(nodeBreakers[r1.Resource]))
		assert.Equal(t, 1, len(currentRules))
		clearData()
	})
}
