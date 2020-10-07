package circuitbreaker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
					Threshold:        10,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.args.rule); got != tt.want {
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
		if got := IsValid(rule); got == nil {
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
		if got := IsValid(rule); got == nil {
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
			Threshold:        0,
		}
		if got := IsValid(rule); got == nil {
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
			Threshold:        10,
		}
		rules = append(rules, r1, r2, r3)
		err := onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, len(breakerRules["abc01"]) == 3)
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
	})

	t.Run("Test_onUpdateRules_invalid", func(t *testing.T) {
		r1 := &Rule{
			Resource: "abc",
		}
		err := onRuleUpdate([]*Rule{r1})
		assert.Nil(t, err)
		assert.True(t, len(GetRules()) == 0)
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
			Threshold:        10,
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
			Threshold:        10,
		}
		r7 := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1100,
			Threshold:        10,
		}
		_, _ = LoadRules([]*Rule{r4, r5, r6, r7})
		assert.True(t, len(breakers) == 1)
		newCbs := breakers["abc"]
		assert.True(t, len(newCbs) == 4, "Expect:4, in fact:", len(newCbs))
		assert.True(t, reflect.DeepEqual(newCbs[0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(newCbs[1].BoundStat(), b2.BoundStat()))
		assert.True(t, reflect.DeepEqual(newCbs[2].BoundRule(), r6))
		assert.True(t, reflect.DeepEqual(newCbs[3].BoundRule(), r7))
	})
}

func TestGeneratorCircuitBreaker(t *testing.T) {
	r := &Rule{
		Resource:         "abc01",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10,
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
			Threshold:        10,
		}

		_, _ = LoadRules([]*Rule{r1})
		rules := GetRules()
		assert.True(t, len(rules) == 1 && rules[0].Resource == r1.Resource && rules[0].Strategy == r1.Strategy)
		_ = ClearRules()
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
	_ = ClearRules()
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
