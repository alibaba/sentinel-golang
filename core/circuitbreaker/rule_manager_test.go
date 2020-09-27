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

func Test_onRuleUpdate_valid(t *testing.T) {
	assert := assert.New(t)

	t.Run("Test_onRuleUpdate_basic", func(t *testing.T) {
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
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
			Resource:         "abc02",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        10,
		}
		rules = append(rules, r1, r2, r3)
		ret, err, failedRules := onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(ret)
		assert.Empty(failedRules)
		assert.Len(breakers["abc01"], 2)
		assert.Len(breakerRules["abc01"], 2)
		assert.Len(breakers["abc02"], 1)
		assert.Len(breakerRules["abc02"], 1)
	})
	t.Run("Test_onRuleUpdate_repeat", func(t *testing.T) {
		var ret bool
		var err error
		var failedRules []*Rule
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		rules := make([]*Rule, 0)
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		rules = append(rules, r)
		_, err, _ = onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		ret, err, failedRules = onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		assert.False(ret)
		assert.Empty(failedRules)
	})
	t.Run("Test_onRuleUpdate_duplicate", func(t *testing.T) {
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		ret, err, failedRules := onRuleUpdate([]*Rule{r, r})
		if err != nil {
			t.Fatal(err)
		}
		assert.True(ret)
		assert.Empty(failedRules)
		assert.Len(breakers["abc"], 1)
		assert.Len(breakerRules["abc"], 1)
	})
	t.Run("Test_onRuleUpdate_remove", func(t *testing.T) {
		var ret bool
		var err error
		var failedRules []*Rule
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		rules := make([]*Rule, 0)
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
		rules = append(rules, r1, r2)
		_, err, _ = onRuleUpdate(rules)
		if err != nil {
			t.Fatal(err)
		}
		ret, err, failedRules = onRuleUpdate(rules[1:])
		if err != nil {
			t.Fatal(err)
		}
		assert.True(ret)
		assert.Empty(failedRules)
		assert.Len(breakers["abc"], 1)
		assert.Len(breakerRules["abc"], 1)
		assert.Equal(breakerRules["abc"][0].Strategy, ErrorRatio)
	})
	t.Run("Test_onRuleUpdate_clear", func(t *testing.T) {
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		onRuleUpdate([]*Rule{r})
		ret, err, failedRules := onRuleUpdate(nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(ret)
		assert.Empty(failedRules)
		assert.Empty(breakers)
		assert.Empty(breakerRules)
	})

	breakers = make(map[string][]CircuitBreaker)
	breakerRules = make(map[string][]*Rule)
}

func Test_onRuleUpdate_invalid(t *testing.T) {
	assert := assert.New(t)

	t.Run("Test_onRuleUpdate_invalid_rule", func(t *testing.T) {
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        -0.3,
		}
		ret, err, failedRules := onRuleUpdate([]*Rule{r})
		if err != nil {
			t.Fatal(err)
		}
		assert.False(ret)
		assert.Len(failedRules, 1)
		assert.Empty(breakers)
		assert.Empty(breakerRules)
	})
	t.Run("Test_onRuleUpdate_unsupported_strategy", func(t *testing.T) {
		breakers = make(map[string][]CircuitBreaker)
		breakerRules = make(map[string][]*Rule)
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        0.3,
		}
		origCbGenFuncMap := cbGenFuncMap
		cbGenFuncMap = make(map[Strategy]CircuitBreakerGenFunc)

		ret, err, failedRules := onRuleUpdate([]*Rule{r})
		if err != nil {
			t.Fatal(err)
		}
		assert.False(ret)
		assert.Len(failedRules, 1)
		assert.Empty(breakers["abc"])
		assert.Empty(breakerRules["abc"])

		cbGenFuncMap = origCbGenFuncMap
	})

	breakers = make(map[string][]CircuitBreaker)
	breakerRules = make(map[string][]*Rule)
}

func TestLoadRules(t *testing.T) {
	assert := assert.New(t)

	t.Run("Test_LoadRules", func(t *testing.T) {
		ClearRules()
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

		_, _, _ = LoadRules([]*Rule{r1, r2, r3})
		b2 := breakers["abc"][1]

		assert.Len(breakers, 1)
		assert.Len(breakers["abc"], 3)
		assert.True(reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))
		assert.True(reflect.DeepEqual(breakers["abc"][1].BoundRule(), r2))
		assert.True(reflect.DeepEqual(breakers["abc"][2].BoundRule(), r3))

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
		_, _, _ = LoadRules([]*Rule{r4, r5, r6, r7})
		assert.Len(breakers, 1)
		newCbs := breakers["abc"]
		assert.Len(newCbs, 4, "Expect:4, in fact:", len(newCbs))
		assert.True(reflect.DeepEqual(newCbs[0].BoundRule(), r1))
		assert.True(reflect.DeepEqual(newCbs[1].BoundStat(), b2.BoundStat()))
		assert.True(reflect.DeepEqual(newCbs[2].BoundRule(), r6))
		assert.True(reflect.DeepEqual(newCbs[3].BoundRule(), r7))
	})
}
