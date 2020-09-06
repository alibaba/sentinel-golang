package circuitbreaker

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isApplicableRule_valid(t *testing.T) {
	type args struct {
		rule Rule
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "rtRule_isApplicable",
			args: args{
				rule: NewRule("abc01", SlowRequestRatio, WithStatIntervalMs(1000),
					WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
					WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1)),
			},
			want: nil,
		},
		{
			name: "errorRatioRule_isApplicable",
			args: args{
				rule: NewRule("abc02", ErrorRatio, WithStatIntervalMs(1000),
					WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
					WithMaxSlowRequestRatio(0.3)),
			},
			want: nil,
		},
		{
			name: "errorCountRule_isApplicable",
			args: args{
				rule: NewRule("abc02", ErrorCount, WithStatIntervalMs(1000),
					WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
					WithMaxSlowRequestRatio(10)),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.rule.IsApplicable(); got != tt.want {
				t.Errorf("RuleManager.IsApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isApplicableRule_invalid(t *testing.T) {
	t.Run("rtBreakerRule_isApplicable_false", func(t *testing.T) {
		rule := NewRule("abc01", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(5),
			WithMinRequestAmount(10050), WithMaxSlowRequestRatio(-1.0))
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
	t.Run("errorRatioRule_isApplicable_false", func(t *testing.T) {
		rule := NewRule("abc02", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithErrorRatioThreshold(-0.3))
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
	t.Run("errorCountRule_isApplicable_false", func(t *testing.T) {
		rule := NewRule("", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0))
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
}

func Test_onUpdateRules(t *testing.T) {
	t.Run("Test_onUpdateRules", func(t *testing.T) {
		rules := make([]Rule, 0)
		r1 := NewRule("abc01", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc01", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc01", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		rules = append(rules, r1, r2, r3)
		err := onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, len(breakers["abc01"]) == 3)
		assert.True(t, len(breakerRules["abc01"]) == 3)
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]Rule)
	})
}

func Test_onRuleUpdate(t *testing.T) {
	t.Run("Test_onRuleUpdate", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		_, _ = LoadRules([]Rule{r1, r2, r3})
		b2 := breakers["abc"][1]

		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(breakers["abc"][1].BoundRule(), r2))
		assert.True(t, reflect.DeepEqual(breakers["abc"][2].BoundRule(), r3))

		r4 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r5 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(100), WithMinRequestAmount(25),
			WithMaxSlowRequestRatio(0.5))
		r6 := NewRule("abc", ErrorCount, WithStatIntervalMs(100),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		r7 := NewRule("abc", ErrorCount, WithStatIntervalMs(1100),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))

		_, _ = LoadRules([]Rule{r4, r5, r6, r7})
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
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1, r2, r3})
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))

		slowRtRuleR1 := r1.(*slowRtRule)
		updateR1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1001),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(30),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.1))
		err = UpdateRule(slowRtRuleR1.Id, updateR1)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), updateR1))
	})

	t.Run("Test_updateRuleReuseStat", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1, r2, r3})
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))

		stat := breakers["abc"][0].BoundStat()

		slowRtRuleR1 := r1.(*slowRtRule)
		updateR1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(21),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.2))
		err = UpdateRule(slowRtRuleR1.Id, updateR1)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundStat(), stat))
	})

	t.Run("Test_notFoundRuleIdError", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1})
		assert.Nil(t, err)

		updateR1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1001),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(30),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.1))
		err = UpdateRule("xxxxxx", updateR1)
		assert.Contains(t, err.Error(), "Rule to be updated was not found,id")
	})

	t.Run("Test_notFoundRuleResourceError", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1})
		assert.Nil(t, err)

		slowRtRuleR1 := r1.(*slowRtRule)
		updateR1 := NewRule("abcd", SlowRequestRatio, WithStatIntervalMs(1001),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(30),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.1))
		err = UpdateRule(slowRtRuleR1.Id, updateR1)
		assert.Contains(t, err.Error(), "Update failed, the current circuitBreaker resource to be updated does not exist")
	})

	t.Run("Test_alreadyExistRuleError", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1})
		assert.Nil(t, err)

		slowRtRuleR1 := r1.(*slowRtRule)
		updateR1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		err = UpdateRule(slowRtRuleR1.Id, updateR1)
		assert.Contains(t, err.Error(), "The rule to be updated already exists.")
	})
}

func Test_appendRule(t *testing.T) {
	t.Run("Test_appendRule", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1, r2, r3})
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))

		r4 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1001),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(30),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.2))
		err = AppendRule(r4)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 4)
		assert.True(t, reflect.DeepEqual(breakers["abc"][3].BoundRule(), r4))
	})

	t.Run("Test_appendRuleByDifferentResource", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		r2 := NewRule("abc", ErrorRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(0.3))
		r3 := NewRule("abc", ErrorCount, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMinRequestAmount(5),
			WithMaxSlowRequestRatio(10))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1, r2, r3})
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))

		r4 := NewRule("abcd", SlowRequestRatio, WithStatIntervalMs(1001),
			WithRetryTimeoutMs(1001), WithMaxAllowedRtMs(30),
			WithMinRequestAmount(6), WithMaxSlowRequestRatio(0.2))
		err = AppendRule(r4)
		assert.Nil(t, err)
		assert.True(t, len(breakers) == 2)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, len(breakers["abcd"]) == 1)
		assert.True(t, reflect.DeepEqual(breakers["abcd"][0].BoundRule(), r4))
	})

	t.Run("Test_alreadyExistRuleError", func(t *testing.T) {
		r1 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		breakers = make(map[string][]CircuitBreaker)
		_, err := LoadRules([]Rule{r1})
		assert.Nil(t, err)

		r2 := NewRule("abc", SlowRequestRatio, WithStatIntervalMs(1000),
			WithRetryTimeoutMs(1000), WithMaxAllowedRtMs(20),
			WithMinRequestAmount(5), WithMaxSlowRequestRatio(0.1))
		err = AppendRule(r2)
		fmt.Println(err)
		assert.Contains(t, err.Error(), "The current appended rule already exists")
	})
}
