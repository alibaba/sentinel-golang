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
	"errors"
	"sync/atomic"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CircuitBreakerMock struct {
	mock.Mock
}

func (m *CircuitBreakerMock) BoundRule() *Rule {
	args := m.Called()
	return args.Get(0).(*Rule)
}

func (m *CircuitBreakerMock) BoundStat() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *CircuitBreakerMock) TryPass(ctx *base.EntryContext) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *CircuitBreakerMock) CurrentState() State {
	args := m.Called()
	return args.Get(0).(State)
}

func (m *CircuitBreakerMock) OnRequestComplete(rt uint64, err error) {
	m.Called(rt, err)
	return
}

type StateChangeListenerMock struct {
	mock.Mock
}

func (s *StateChangeListenerMock) OnTransformToClosed(prev State, rule Rule) {
	logging.Debug("transform to closed", "strategy", rule.Strategy, "prevState", prev.String())
	return
}

func (s *StateChangeListenerMock) OnTransformToOpen(prev State, rule Rule, snapshot interface{}) {
	logging.Debug("transform to open", "strategy", rule.Strategy, "prevState", prev.String(), "snapshot", snapshot)
}

func (s *StateChangeListenerMock) OnTransformToHalfOpen(prev State, rule Rule) {
	logging.Debug("transform to Half-Open", "strategy", rule.Strategy, "prevState", prev.String())
}

func TestStatus(t *testing.T) {
	t.Run("get_set", func(t *testing.T) {
		status := newState()
		assert.True(t, status.get() == Closed)

		status.set(Open)
		assert.True(t, status.get() == Open)
	})

	t.Run("cas", func(t *testing.T) {
		status := newState()
		assert.True(t, status.get() == Closed)

		assert.True(t, status.cas(Closed, Open))
		assert.True(t, !status.cas(Closed, Open))
		status.set(HalfOpen)
		assert.True(t, status.cas(HalfOpen, Open))
	})
}

func TestSlowRtCircuitBreaker_TryPass(t *testing.T) {
	ClearStateChangeListeners()
	stateChangeListenerMock := &StateChangeListenerMock{}
	stateChangeListenerMock.On("OnTransformToHalfOpen", Open, mock.Anything).Return()
	RegisterStateChangeListeners(stateChangeListenerMock)
	t.Run("TryPass_Closed", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   50,
			Threshold:        0.5,
		}
		b, err := newSlowRtCircuitBreaker(r)
		assert.Nil(t, err)
		pass := b.TryPass(base.NewEmptyEntryContext())
		assert.True(t, pass)
	})

	t.Run("TryPass_Probe", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   50,
			Threshold:        0.5,
		}
		b, err := newSlowRtCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		pass := b.TryPass(ctx)
		assert.True(t, pass)
		assert.True(t, b.state.get() == HalfOpen)
	})

	t.Run("TryPass_ProbeNum", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   50,
			Threshold:        0.5,
			ProbeNum:         10,
		}
		b, err := newSlowRtCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		for i := 0; i < 10; i++ {
			pass := b.TryPass(ctx)
			assert.True(t, pass)
			assert.True(t, b.state.get() == HalfOpen)
			b.OnRequestComplete(1, nil)
		}
		assert.True(t, b.state.get() == Closed)
	})
}

func TestSlowRt_OnRequestComplete(t *testing.T) {
	ClearStateChangeListeners()
	r := &Rule{
		Resource:         "abc",
		Strategy:         SlowRequestRatio,
		RetryTimeoutMs:   3000,
		MinRequestAmount: 10,
		StatIntervalMs:   10000,
		MaxAllowedRtMs:   50,
		Threshold:        0.5,
	}
	b, err := newSlowRtCircuitBreaker(r)
	assert.Nil(t, err)
	t.Run("OnRequestComplete_Less_Than_MinRequestMount", func(t *testing.T) {
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_Probe_Failed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Open)
	})
	t.Run("OnRequestComplete_Probe_Succeed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(10, nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_ProbeNum_Success", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(10, nil)
		assert.True(t, b.CurrentState() == HalfOpen)
		assert.True(t, b.curProbeNumber == 1)
	})
	t.Run("OnRequestComplete_ProbeNum_Failed", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Open)
		assert.True(t, b.curProbeNumber == 0)
	})
}

func TestSlowRt_ResetBucketTo(t *testing.T) {
	t.Run("ResetBucketTo", func(t *testing.T) {
		wrap := &sbase.BucketWrap{
			BucketStart: 1,
			Value:       atomic.Value{},
		}
		wrap.Value.Store(&slowRequestCounter{
			slowCount:  1,
			totalCount: 1,
		})

		la := &slowRequestLeapArray{}
		la.ResetBucketTo(wrap, util.CurrentTimeMillis())
		counter := wrap.Value.Load().(*slowRequestCounter)
		assert.True(t, counter.totalCount == 0 && counter.slowCount == 0)
	})
}

func TestErrorRatioCircuitBreaker_TryPass(t *testing.T) {
	t.Run("TryPass_Closed", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        0.5,
		}
		b, err := newErrorRatioCircuitBreaker(r)
		assert.Nil(t, err)
		pass := b.TryPass(base.NewEmptyEntryContext())
		assert.True(t, pass)
	})

	t.Run("TryPass_Probe", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        0.5,
		}
		b, err := newErrorRatioCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		pass := b.TryPass(ctx)
		assert.True(t, pass)
		assert.True(t, b.state.get() == HalfOpen)
	})
	t.Run("TryPass_ProbeNum", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        0.5,
			ProbeNum:         10,
		}
		b, err := newErrorRatioCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		for i := 0; i < 10; i++ {
			pass := b.TryPass(ctx)
			assert.True(t, pass)
			assert.True(t, b.state.get() == HalfOpen)
			b.OnRequestComplete(1, nil)
		}
		assert.True(t, b.state.get() == Closed)
	})
}

func TestErrorRatio_OnRequestComplete(t *testing.T) {
	r := &Rule{
		Resource:         "abc",
		Strategy:         ErrorRatio,
		RetryTimeoutMs:   3000,
		MinRequestAmount: 10,
		StatIntervalMs:   10000,
		Threshold:        0.5,
	}
	b, err := newErrorRatioCircuitBreaker(r)
	assert.Nil(t, err)
	t.Run("OnRequestComplete_Less_Than_MinRequestAmount", func(t *testing.T) {
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_Probe_Succeed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_Probe_Failed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(0, errors.New("errorRatio"))
		assert.True(t, b.CurrentState() == Open)
	})
	t.Run("OnRequestComplete_ProbeNum_Success", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == HalfOpen)
		assert.True(t, b.curProbeNumber == 1)
	})
	t.Run("OnRequestComplete_ProbeNum_Failed", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(0, errors.New("errorRatio"))
		assert.True(t, b.CurrentState() == Open)
		assert.True(t, b.curProbeNumber == 0)
	})
}

func TestErrorRatio_ResetBucketTo(t *testing.T) {
	t.Run("ResetBucketTo", func(t *testing.T) {
		wrap := &sbase.BucketWrap{
			BucketStart: 1,
			Value:       atomic.Value{},
		}
		wrap.Value.Store(&errorCounter{
			errorCount: 1,
			totalCount: 1,
		})

		la := &errorCounterLeapArray{}
		la.ResetBucketTo(wrap, util.CurrentTimeMillis())
		counter := wrap.Value.Load().(*errorCounter)
		assert.True(t, counter.errorCount == 0 && counter.totalCount == 0)
	})
}

func TestErrorCountCircuitBreaker_TryPass(t *testing.T) {
	t.Run("TryPass_Closed", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        1.0,
		}
		b, err := newErrorCountCircuitBreaker(r)
		assert.Nil(t, err)
		pass := b.TryPass(base.NewEmptyEntryContext())
		assert.True(t, pass)
	})

	t.Run("TryPass_Probe", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        1.0,
		}
		b, err := newErrorCountCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		pass := b.TryPass(ctx)
		assert.True(t, pass)
		assert.True(t, b.state.get() == HalfOpen)
	})

	t.Run("TryPass_ProbeNum", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        1.0,
			ProbeNum:         10,
		}
		b, err := newErrorCountCircuitBreaker(r)
		assert.Nil(t, err)

		b.state.set(Open)
		ctx := &base.EntryContext{
			Resource: base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound),
		}
		e := base.NewSentinelEntry(ctx, base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound), nil)
		ctx.SetEntry(e)
		for i := 0; i < 10; i++ {
			pass := b.TryPass(ctx)
			assert.True(t, pass)
			assert.True(t, b.state.get() == HalfOpen)
			b.OnRequestComplete(1, nil)
		}
		assert.True(t, b.state.get() == Closed)
	})
}

func TestErrorCount_OnRequestComplete(t *testing.T) {
	r := &Rule{
		Resource:         "abc",
		Strategy:         ErrorCount,
		RetryTimeoutMs:   3000,
		MinRequestAmount: 10,
		StatIntervalMs:   10000,
		Threshold:        1.0,
	}
	b, err := newErrorCountCircuitBreaker(r)
	assert.Nil(t, err)
	t.Run("OnRequestComplete_Less_Than_MinRequestAmount", func(t *testing.T) {
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_Probe_Succeed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == Closed)
	})
	t.Run("OnRequestComplete_Probe_Failed", func(t *testing.T) {
		b.state.set(HalfOpen)
		b.OnRequestComplete(0, errors.New("errorCount"))
		assert.True(t, b.CurrentState() == Open)
	})
	t.Run("OnRequestComplete_ProbeNum_Success", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(base.NewEmptyEntryContext().Rt(), nil)
		assert.True(t, b.CurrentState() == HalfOpen)
		assert.True(t, b.curProbeNumber == 1)
	})
	t.Run("OnRequestComplete_ProbeNum_Failed", func(t *testing.T) {
		b.probeNumber = 2
		b.state.set(HalfOpen)
		b.OnRequestComplete(0, errors.New("errorCount"))
		assert.True(t, b.CurrentState() == Open)
		assert.True(t, b.curProbeNumber == 0)
	})
}

func TestFromClosedToOpen(t *testing.T) {
	ClearStateChangeListeners()
	stateChangeListenerMock := &StateChangeListenerMock{}
	stateChangeListenerMock.On("OnTransformToOpen", Closed, mock.Anything, mock.Anything).Return()
	RegisterStateChangeListeners(stateChangeListenerMock)
	t.Run("FromCloseToOpen", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        1.0,
		}
		b, err := newErrorCountCircuitBreaker(r)
		assert.Nil(t, err)
		statusChanged := b.fromClosedToOpen("")
		assert.True(t, statusChanged)
		stateChangeListenerMock.MethodCalled("OnTransformToOpen", Closed, mock.Anything, mock.Anything)
	})
}

func TestFromHalfOpenToOpen(t *testing.T) {
	ClearStateChangeListeners()
	stateChangeListenerMock := &StateChangeListenerMock{}
	stateChangeListenerMock.On("OnTransformToOpen", HalfOpen, mock.Anything, mock.Anything).Return()
	RegisterStateChangeListeners(stateChangeListenerMock)
	t.Run("FromHalfOpenToOpen", func(t *testing.T) {
		r := &Rule{
			Resource:         "abc",
			Strategy:         ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   10000,
			Threshold:        1.0,
		}
		b, err := newErrorCountCircuitBreaker(r)
		assert.Nil(t, err)
		b.state.set(HalfOpen)
		statusChanged := b.fromHalfOpenToOpen("")
		assert.True(t, statusChanged)
		assert.True(t, b.nextRetryTimestampMs > 0)
		stateChangeListenerMock.MethodCalled("OnTransformToOpen", HalfOpen, mock.Anything, mock.Anything)
	})
}
