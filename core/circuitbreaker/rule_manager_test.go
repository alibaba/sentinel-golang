package circuitbreaker

import (
	"fmt"
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

func Test_updateSpecifiedRule(t *testing.T) {
	t.Run("Test_updateSpecifiedRule", func(t *testing.T) {
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
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundRule(), r1))

		updateR1 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   2000,
			MinRequestAmount: 6,
			StatIntervalMs:   2000,
			MaxAllowedRtMs:   30,
			Threshold:        0.2,
		}
		err = UpdateRule(r1.Id, updateR1)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundRule(), updateR1))
	})

	t.Run("Test_updateRuleReuseStat", func(t *testing.T) {
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
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundRule(), r1))

		stat := breakers["abc01"][0].BoundStat()

		updateR1 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   10001,
			MinRequestAmount: 6,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   21,
			Threshold:        0.2,
		}
		err = UpdateRule(r1.Id, updateR1)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundStat(), stat))
	})

	t.Run("Test_notFoundRuleIdError", func(t *testing.T) {
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
		rules = append(rules, r1)
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)

		updateR1 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1001,
			MinRequestAmount: 6,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   21,
			Threshold:        0.2,
		}
		err = UpdateRule("xxxxxx", updateR1)
		assert.Contains(t, err.Error(), "Rule to be updated was not found,id")
	})

	t.Run("Test_notFoundRuleResourceError", func(t *testing.T) {
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
		rules = append(rules, r1)
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)

		updateR1 := &Rule{
			Resource:         "abc012",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		err = UpdateRule(r1.Id, updateR1)
		assert.Contains(t, err.Error(), "Update failed, the current circuitBreaker resource to be updated does not exist")
	})

	t.Run("Test_alreadyExistRuleError", func(t *testing.T) {
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
		rules = append(rules, r1)
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)

		updateR1 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		err = UpdateRule(r1.Id, updateR1)
		assert.Contains(t, err.Error(), "The rule to be updated already exists.")
	})
}

func Test_appendRule(t *testing.T) {
	t.Run("Test_appendRule", func(t *testing.T) {
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
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundRule(), r1))

		r4 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1001,
			MinRequestAmount: 6,
			StatIntervalMs:   1001,
			MaxAllowedRtMs:   30,
			Threshold:        0.2,
		}
		err = AppendRule(r4)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc01"]) == 4)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][3].BoundRule(), r4))
	})

	t.Run("Test_appendRuleByDifferentResource", func(t *testing.T) {
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
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc01"][0].BoundRule(), r1))

		r4 := &Rule{
			Resource:         "abc02",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1001,
			MinRequestAmount: 6,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   30,
			Threshold:        0.2,
		}
		err = AppendRule(r4)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 2)
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, len(breakers["abc02"]) == 1)
		assert.True(t, reflect.DeepEqual(breakers["abc02"][0].BoundRule(), r4))
	})

	t.Run("Test_alreadyExistRuleError", func(t *testing.T) {
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
		rules = append(rules, r1)
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules(rules)
		assert.Nil(t, err)

		r2 := &Rule{
			Resource:         "abc01",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		err = AppendRule(r2)
		fmt.Println(err)
		assert.Contains(t, err.Error(), "The current appended rule already exists")
	})
}
