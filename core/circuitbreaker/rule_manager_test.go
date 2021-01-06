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

package circuitbreaker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func clearData() {
	breakerRules = make(map[string][]*Rule)
	breakers = make(map[string][]CircuitBreaker)
	currentRules = make(map[string][]*Rule, 0)
}

func Test_isApplicableRule_valid(t *testing.T) {
	type args struct {
		rule *Rule
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "rtRule_isApplicable",
			args: args{
				rule: &Rule{
					Resource:         "abc01",
					Strategy:         SlowRequestRatio,
					RetryTimeoutMs:   1000,
					MinRequestAmount: 5,
					StatIntervalMs:   1000,
					MaxAllowedRtMs:   20,
					Threshold:        0.1,
				},
			},
			want: nil,
		},
		{
			name: "errorRatioRule_isApplicable",
			args: args{
				rule: &Rule{
					Resource:         "abc02",
					Strategy:         ErrorRatio,
					RetryTimeoutMs:   1000,
					MinRequestAmount: 5,
					StatIntervalMs:   1000,
					Threshold:        0.3,
				},
			},
			want: nil,
		},
		{
			name: "errorCountRule_isApplicable",
			args: args{
				rule: &Rule{
					Resource:         "abc02",
					Strategy:         ErrorCount,
					RetryTimeoutMs:   1000,
					MinRequestAmount: 5,
					StatIntervalMs:   1000,
					Threshold:        10.0,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRule(tt.args.rule); got != tt.want {
				t.Errorf("RuleManager.isApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isApplicableRule_invalid(t *testing.T) {
	t.Run("rtBreakerRule_isApplicable_false", func(t *testing.T) {
		rule := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 10050,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   5,
			Threshold:        -1.0,
		}
		if got := IsValidRule(rule); got == nil {
			t.Errorf("RuleManager.isApplicable() = %v", got)
		}
	})
	t.Run("errorRatioRule_isApplicable_false", func(t *testing.T) {
		rule := &Rule{
			Resource:         "abc02",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        -0.3,
		}
		if got := IsValidRule(rule); got == nil {
			t.Errorf("RuleManager.isApplicable() = %v", got)
		}
	})
	t.Run("errorCountRule_isApplicable_false", func(t *testing.T) {
		rule := &Rule{
			Resource:         "",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        0.0,
		}
		if got := IsValidRule(rule); got == nil {
			t.Errorf("RuleManager.isApplicable() = %v", got)
		}
	})
}

func Test_onUpdateRules(t *testing.T) {
	t.Run("Test_onUpdateRules", func(t *testing.T) {
		rules := make([]*Rule, 0)
		r1 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		r2 := &Rule{
			Resource:         "abc01",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        0.3,
		}
		r3 := &Rule{
			Resource:         "abc01",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        10.0,
		}
		rules = append(rules, r1, r2, r3)
		resRulesMap := make(map[string][]*Rule)
		resRulesMap["abc01"] = rules
		err := onRuleUpdate(resRulesMap)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, len(breakerRules["abc01"]) == 3)
		clearData()
	})

	t.Run("Test_onUpdateRules_invalid", func(t *testing.T) {
		r1 := &Rule{
			Resource: "abc",
		}
		resRulesMap := make(map[string][]*Rule)
		resRulesMap["abc01"] = []*Rule{r1}
		err := onRuleUpdate(resRulesMap)
		assert.Nil(t, err)
		assert.True(t, len(GetRules()) == 0)
		clearData()
	})
}

func Test_onRuleUpdate(t *testing.T) {
	t.Run("Test_onRuleUpdate", func(t *testing.T) {
		r1 := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		r2 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        0.3,
		}
		r3 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        10.0,
		}

		_, _ = LoadRules([]*Rule{r1, r2, r3})
		b2 := breakers["abc"][1]

		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(breakers["abc"][1].BoundRule(), r2))
		assert.True(t, reflect.DeepEqual(breakers["abc"][2].BoundRule(), r3))

		r4 := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		r5 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   100,
			MinRequestAmount: 25,
			StatIntervalMs:   1000,
			Threshold:        0.5,
		}
		r6 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   100,
			Threshold:        10.0,
		}
		r7 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1100,
			Threshold:        10.0,
		}
		_, _ = LoadRules([]*Rule{r4, r5, r6, r7})
		assert.True(t, len(breakers) == 1)
		newCbs := breakers["abc"]
		assert.True(t, len(newCbs) == 4, "Expect:4, in fact:", len(newCbs))
		assert.True(t, reflect.DeepEqual(newCbs[0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(newCbs[1].BoundStat(), b2.BoundStat()))
		assert.True(t, reflect.DeepEqual(newCbs[2].BoundRule(), r6))
		assert.True(t, reflect.DeepEqual(newCbs[3].BoundRule(), r7))
		clearData()
	})
}

func TestGeneratorCircuitBreaker(t *testing.T) {
	r := &Rule{
		Resource:         "abc01",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10.0,
	}
	t.Run("SlowRequestRatio_Nil_Rule", func(t *testing.T) {
		generator := cbGenFuncMap[SlowRequestRatio]
		cb, err := generator(nil, nil)
		assert.Nil(t, cb)
		assert.Error(t, err, "nil rule")
	})

	t.Run("SlowRequestRatio_ReuseStat_Unmatched_Type", func(t *testing.T) {
		generator := cbGenFuncMap[SlowRequestRatio]
		_, err := generator(r, &errorCounterLeapArray{})
		assert.Nil(t, err)
	})

	t.Run("ErrorRatio_ReuseStat_Unmatched_Type", func(t *testing.T) {
		generator := cbGenFuncMap[ErrorRatio]
		_, err := generator(r, &slowRequestLeapArray{})
		assert.Nil(t, err)
	})

	t.Run("ErrorCount_ReuseStat_Unmatched_Type", func(t *testing.T) {
		generator := cbGenFuncMap[ErrorCount]
		_, err := generator(r, &slowRequestLeapArray{})
		assert.Nil(t, err)
	})
}

func TestGetRules(t *testing.T) {
	t.Run("TestGetRules", func(t *testing.T) {
		r1 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        10.0,
		}

		_, _ = LoadRules([]*Rule{r1})
		rules := GetRules()
		assert.True(t, len(rules) == 1 && rules[0].Resource == r1.Resource && rules[0].Strategy == r1.Strategy)
		clearData()
	})
}

func TestGetBreakersOfResource(t *testing.T) {
	r1 := &Rule{
		Resource:         "abc",
		Strategy:         SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   20,
		Threshold:        0.1,
	}

	_, _ = LoadRules([]*Rule{r1})

	cbs := getBreakersOfResource("abc")
	assert.True(t, len(cbs) == 1 && cbs[0].BoundRule() == r1)
	clearData()
}

func TestSetCircuitBreakerGenerator(t *testing.T) {
	t.Run("TestSetCircuitBreakerGenerator_Normal", func(t *testing.T) {
		err := SetCircuitBreakerGenerator(100, func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
			return newSlowRtCircuitBreakerWithStat(r, nil), nil
		})
		assert.Nil(t, err)
	})

	t.Run("TestSetCircuitBreakerGenerator_Err", func(t *testing.T) {
		err := SetCircuitBreakerGenerator(SlowRequestRatio, func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
			return newSlowRtCircuitBreakerWithStat(r, nil), nil
		})
		assert.Error(t, err, "not allowed to replace the generator for default circuit breaking strategies")
	})
}

func TestRemoveCircuitBreakerGenerator(t *testing.T) {
	t.Run("TestRemoveCircuitBreakerGenerator_Normal", func(t *testing.T) {
		err := RemoveCircuitBreakerGenerator(100)
		assert.Nil(t, err)
	})

	t.Run("TestRemoveCircuitBreakerGenerator_Err", func(t *testing.T) {
		err := RemoveCircuitBreakerGenerator(SlowRequestRatio)
		assert.Error(t, err, "not allowed to remove the generator for default circuit breaking strategies")
	})
}

func TestLoadRules(t *testing.T) {
	t.Run("loadSameRules", func(t *testing.T) {
		_, err := LoadRules([]*Rule{
			{
				Resource:         "abc",
				Strategy:         SlowRequestRatio,
				RetryTimeoutMs:   1000,
				MinRequestAmount: 5,
				StatIntervalMs:   1000,
				MaxAllowedRtMs:   20,
				Threshold:        0.1,
			},
		})
		assert.Nil(t, err)
		ok, err := LoadRules([]*Rule{
			{
				Resource:         "abc",
				Strategy:         SlowRequestRatio,
				RetryTimeoutMs:   1000,
				MinRequestAmount: 5,
				StatIntervalMs:   1000,
				MaxAllowedRtMs:   20,
				Threshold:        0.1,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
		clearData()
	})
}

func TestLoadRulesOfResource(t *testing.T) {
	r1 := &Rule{
		Resource:         "abc1",
		Strategy:         SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   20,
		Threshold:        0.1,
	}
	r2 := &Rule{
		Resource:         "abc1",
		Strategy:         ErrorRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        0.3,
	}
	r3 := &Rule{
		Resource:         "abc2",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10.0,
	}
	succ, err := LoadRules([]*Rule{r1, r2, r3})
	assert.True(t, succ && err == nil)

	t.Run("LoadRulesOfResource_empty_resource", func(t *testing.T) {
		succ, err = LoadRulesOfResource("", []*Rule{r1, r2})
		assert.True(t, !succ && err != nil)
	})

	t.Run("LoadRulesOfResource_cache_hit", func(t *testing.T) {
		r11 := *r1
		r12 := *r2
		succ, err = LoadRulesOfResource("abc1", []*Rule{&r11, &r12})
		assert.True(t, !succ && err == nil)
	})

	t.Run("LoadRulesOfResource_clear", func(t *testing.T) {
		succ, err = LoadRulesOfResource("abc1", []*Rule{})
		assert.True(t, succ && err == nil)
		assert.True(t, len(breakerRules["abc1"]) == 0 && len(currentRules["abc1"]) == 0)
		assert.True(t, len(breakerRules["abc2"]) == 1 && len(currentRules["abc2"]) == 1)
	})
	clearData()
}

func Test_onResourceRuleUpdate(t *testing.T) {
	r1 := Rule{
		Resource:         "abc1",
		Strategy:         SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   20,
		Threshold:        0.1,
	}
	r2 := Rule{
		Resource:         "abc1",
		Strategy:         ErrorRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        0.3,
	}
	r3 := Rule{
		Resource:         "abc2",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10.0,
	}
	succ, err := LoadRules([]*Rule{&r1, &r2, &r3})
	assert.True(t, succ && err == nil)

	t.Run("Test_onResourceRuleUpdate_normal", func(t *testing.T) {
		r11 := r1
		r11.Threshold = 0.5
		err = onResourceRuleUpdate("abc1", []*Rule{&r11})

		assert.True(t, len(breakerRules["abc1"]) == 1)
		assert.True(t, len(breakers["abc1"]) == 1)
		assert.True(t, len(currentRules["abc1"]) == 1)
		assert.True(t, breakers["abc1"][0].BoundRule() == &r11)

		assert.True(t, len(breakerRules["abc2"]) == 1)
		assert.True(t, len(breakers["abc2"]) == 1)
		assert.True(t, len(currentRules["abc2"]) == 1)

		clearData()
	})
}

func TestClearRulesOfResource(t *testing.T) {
	r1 := Rule{
		Resource:         "abc1",
		Strategy:         SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   20,
		Threshold:        0.1,
	}
	r2 := Rule{
		Resource:         "abc1",
		Strategy:         ErrorRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        0.3,
	}
	r3 := Rule{
		Resource:         "abc2",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10.0,
	}
	succ, err := LoadRules([]*Rule{&r1, &r2, &r3})
	assert.True(t, succ && err == nil)

	t.Run("TestClearRulesOfResource_normal", func(t *testing.T) {
		assert.True(t, ClearRulesOfResource("abc1") == nil)

		assert.True(t, len(breakerRules["abc1"]) == 0)
		assert.True(t, len(breakers["abc1"]) == 0)
		assert.True(t, len(currentRules["abc1"]) == 0)
		assert.True(t, len(breakerRules["abc2"]) == 1)
		assert.True(t, len(breakers["abc2"]) == 1)
		assert.True(t, len(currentRules["abc2"]) == 1)
		clearData()
	})
}
