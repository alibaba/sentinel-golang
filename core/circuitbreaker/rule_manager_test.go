package circuitbreaker

import (
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
				rule: NewSlowRtRule("abc01", 1000, 1, 20, 5, 0.1),
			},
			want: nil,
		},
		{
			name: "errorRatioRule_isApplicable",
			args: args{
				rule: NewErrorRatioRule("abc02", 1000, 1, 5, 0.3),
			},
			want: nil,
		},
		{
			name: "errorCountRule_isApplicable",
			args: args{
				rule: NewErrorCountRule("abc03", 1000, 1, 5, 10),
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
		rule := NewSlowRtRule("abc01", 1000, 1, 5, 10050, -1.0)
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
	t.Run("errorRatioRule_isApplicable_false", func(t *testing.T) {
		rule := NewErrorRatioRule("abc02", 1000, 1, 5, -0.3)
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
	t.Run("errorCountRule_isApplicable_false", func(t *testing.T) {
		rule := NewErrorCountRule("", 1000, 1, 5, 0)
		if got := rule.IsApplicable(); got == nil {
			t.Errorf("RuleManager.IsApplicable() = %v", got)
		}
	})
}

func Test_onUpdateRules(t *testing.T) {
	t.Run("Test_onUpdateRules", func(t *testing.T) {
		rules := make([]Rule, 0)
		rules = append(rules, NewSlowRtRule("abc01", 1000, 1, 20, 5, 0.1))
		rules = append(rules, NewErrorRatioRule("abc01", 1000, 1, 5, 0.3))
		rules = append(rules, NewErrorCountRule("abc01", 1000, 1, 5, 10))
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
		r1 := NewSlowRtRule("abc", 1000, 1, 20, 5, 0.1)
		r2 := NewErrorRatioRule("abc", 1000, 1, 5, 0.3)
		r3 := NewErrorCountRule("abc", 1000, 1, 5, 10)
		_, _ = LoadRules([]Rule{r1, r2, r3})
		b2 := breakers["abc"][1]

		assert.True(t, len(breakers) == 1)
		assert.True(t, len(breakers["abc"]) == 3)
		assert.True(t, reflect.DeepEqual(breakers["abc"][0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(breakers["abc"][1].BoundRule(), r2))
		assert.True(t, reflect.DeepEqual(breakers["abc"][2].BoundRule(), r3))

		r4 := NewSlowRtRule("abc", 1000, 1, 20, 5, 0.1)
		r5 := NewErrorRatioRule("abc", 1000, 100, 25, 0.5)
		r6 := NewErrorCountRule("abc", 100, 1, 5, 10)
		r7 := NewErrorCountRule("abc", 1100, 1, 5, 10)

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
