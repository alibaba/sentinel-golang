package circuitbreaker

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Check(t *testing.T) {

	t.Run("Test_Custom_CircuitBreaker_Strategy_Check", func(t *testing.T) {
		rules := []*Rule{
			{
				Resource:         "abc",
				Strategy:         101,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 10,
				StatIntervalMs:   10000,
				MaxAllowedRtMs:   50,
				Threshold:        0.5,
			},
		}
		e := SetCircuitBreakerGenerator(101, func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
			circuitBreakerMock := &CircuitBreakerMock{}
			circuitBreakerMock.On("TryPass", mock.Anything).Return(false)
			circuitBreakerMock.On("BoundRule", mock.Anything).Return(rules[0])
			return circuitBreakerMock, nil
		})
		assert.True(t, e == nil)

		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(getBreakersOfResource("abc")) == 1)
		s := &Slot{}
		ctx := &base.EntryContext{
			Resource:        base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
			RuleCheckResult: base.NewTokenResultPass(),
		}
		token := s.Check(ctx)
		assert.True(t, token.IsBlocked())
		_ = ClearRules()
	})

	t.Run("TestCheck_NoPass_NewTokenResultBlocked", func(t *testing.T) {
		rules := []*Rule{
			{
				Resource:         "abc",
				Strategy:         102,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 10,
				StatIntervalMs:   10000,
				MaxAllowedRtMs:   50,
				Threshold:        0.5,
			},
		}
		e := SetCircuitBreakerGenerator(102, func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
			circuitBreakerMock := &CircuitBreakerMock{}
			circuitBreakerMock.On("TryPass", mock.Anything).Return(false)
			circuitBreakerMock.On("BoundRule", mock.Anything).Return(rules[0])
			return circuitBreakerMock, nil
		})
		assert.True(t, e == nil)

		_, err := LoadRules(rules)
		assert.Nil(t, err)
		assert.True(t, len(getBreakersOfResource("abc")) == 1)

		s := &Slot{}
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		token := s.Check(ctx)
		assert.True(t, token.IsBlocked())
		_ = ClearRules()
	})

	t.Run("TestCheck_Pass", func(t *testing.T) {
		e := SetCircuitBreakerGenerator(100, func(r *Rule, reuseStat interface{}) (CircuitBreaker, error) {
			circuitBreakerMock := &CircuitBreakerMock{}
			circuitBreakerMock.On("TryPass", mock.Anything).Return(true)
			return circuitBreakerMock, nil
		})
		assert.True(t, e == nil)

		_, err := LoadRules([]*Rule{
			{
				Resource:         "abc",
				Strategy:         100,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 10,
				StatIntervalMs:   10000,
				MaxAllowedRtMs:   50,
				Threshold:        0.5,
			},
		})
		assert.Nil(t, err)
		assert.True(t, len(getBreakersOfResource("abc")) == 1)

		s := &Slot{}
		ctx := &base.EntryContext{
			Resource:        base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
			RuleCheckResult: base.NewTokenResultPass(),
		}
		token := s.Check(ctx)
		assert.True(t, reflect.DeepEqual(token, ctx.RuleCheckResult))
		_ = ClearRules()
	})

	t.Run("TestCheck_No_Resource", func(t *testing.T) {
		s := &Slot{}
		ctx := &base.EntryContext{
			Resource:        base.NewResourceWrapper("", base.ResTypeCommon, base.Inbound),
			RuleCheckResult: base.NewTokenResultPass(),
		}
		token := s.Check(ctx)
		assert.True(t, reflect.DeepEqual(token, ctx.RuleCheckResult))
	})
}
