package circuit_breaker

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_isApplicableRule(t *testing.T) {
	type args struct {
		rule Rule
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "averageRtBreakerRule_isApplicable",
			args: args{
				rule: NewAverageRtRule("abc01", 1, 2, 1000, 100, 5),
			},
			want: true,
		},
		{
			name: "errorRatioBreakerRule_isApplicable",
			args: args{
				rule: NewErrorRatioRule("abc02", 1, 2, 1000, 0.3, 5),
			},
			want: true,
		},
		{
			name: "errorCountBreakerRule_isApplicable",
			args: args{
				rule: NewErrorCountRule("abc03", 1, 2, 1000, 10),
			},
			want: true,
		},
		{
			name: "averageRtBreakerRule_isApplicable_false",
			args: args{
				rule: NewAverageRtRule("abc01", 1, 2, 1000, -1.0, 5),
			},
			want: false,
		},
		{
			name: "errorRatioBreakerRule_isApplicable_false",
			args: args{
				rule: NewErrorRatioRule("abc02", 1, 2, 1000, -0.3, 5),
			},
			want: false,
		},
		{
			name: "errorCountBreakerRule_isApplicable_false",
			args: args{
				rule: NewErrorCountRule("abc03", 1, 2, 1000, -10),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.rule.isApplicable(); got != tt.want {
				t.Errorf("RuleManager.isApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_onUpdateRules(t *testing.T) {
	type args struct {
		rules []Rule
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_onUpdateRules",
			args: args{
				rules: make([]Rule, 0, 3),
			},
		},
	}

	averageRtRule := &ruleMock{}
	averageRtBreaker := &circuitBreakerMock{}
	averageRtRule.On("isApplicable").Return(true)
	averageRtRule.On("BreakerStrategy").Return(AverageRt)
	averageRtRule.On("String").Return("averageRtRule")
	averageRtRule.On("ResourceName").Return("a")
	averageRtRule.On("convert2CircuitBreaker").Return(averageRtBreaker)
	averageRtBreaker.On("getRule").Return(averageRtRule)
	tests[0].args.rules = append(tests[0].args.rules, averageRtRule)

	errRatioRule := &ruleMock{}
	errRatioBreaker := &circuitBreakerMock{}
	errRatioRule.On("isApplicable").Return(true)
	errRatioRule.On("BreakerStrategy").Return(ErrorRatio)
	errRatioRule.On("String").Return("errRatioRule")
	errRatioRule.On("ResourceName").Return("a")
	errRatioRule.On("convert2CircuitBreaker").Return(errRatioBreaker)
	errRatioBreaker.On("getRule").Return(errRatioRule)
	tests[0].args.rules = append(tests[0].args.rules, errRatioRule)

	errCountRule := &ruleMock{}
	errCountBreaker := &circuitBreakerMock{}
	errCountRule.On("isApplicable").Return(true)
	errCountRule.On("BreakerStrategy").Return(ErrorCount)
	errCountRule.On("String").Return("errCountRule")
	errCountRule.On("ResourceName").Return("a")
	errCountRule.On("convert2CircuitBreaker").Return(errCountBreaker)
	errCountBreaker.On("getRule").Return(errCountRule)
	tests[0].args.rules = append(tests[0].args.rules, errCountRule)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = onRuleUpdate(tt.args.rules)
			assert.True(t, len(breakers["a"]) == 3)
			assert.True(t, len(breakerRules["a"]) == 3)
			for idx, breaker := range breakers["a"] {
				reflect.DeepEqual(breaker.getRule(), tests[0].args.rules[idx])
			}
		})
	}
}
