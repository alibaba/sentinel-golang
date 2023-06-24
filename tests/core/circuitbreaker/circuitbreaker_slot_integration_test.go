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
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StateChangeListenerMock struct {
	mock.Mock
}

func (s *StateChangeListenerMock) OnTransformToClosed(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	_ = s.Called(prev, rule)
	logging.Debug("transform to closed", "strategy", rule.Strategy, "prevState", prev.String())
	return
}

func (s *StateChangeListenerMock) OnTransformToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	_ = s.Called(prev, rule, snapshot)
	logging.Debug("transform to open", "strategy", rule.Strategy, "prevState", prev.String(), "snapshot", snapshot)
}

func (s *StateChangeListenerMock) OnTransformToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	_ = s.Called(prev, rule)
	logging.Debug("transform to Half-Open", "strategy", rule.Strategy, "prevState", prev.String())
}

// Test scenario
// circuit breaker1: slow rt, max rt: 3ms, retry timeout: 1ms, slowRt threshold: 0.1
// circuit breaker2: error ratio, retry timeout: 2000000+ms, error ratio threshold: 0.1
// First request: make cb1 and cb2 trigger fusing
// Second request: make cb1 retry and change state from open to halfOpen, but this request is blocked by cb2.
//
//	when request exit, rollback the state of cb1 to open
//
// Third request: same with second request.
func TestCircuitBreakerSlotIntegration_Normal(t *testing.T) {
	util.SetClock(util.NewMockClock())

	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}

	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		t.Fatal(err)
	}

	cbRule1 := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   1,
		MinRequestAmount: 0,
		StatIntervalMs:   10000,
		MaxAllowedRtMs:   3,
		Threshold:        0.1,
	}
	cbRule2 := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.ErrorRatio,
		RetryTimeoutMs:   2000000,
		MinRequestAmount: 0,
		StatIntervalMs:   10000,
		Threshold:        0.1,
	}

	_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{cbRule1, cbRule2})
	stateListener := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener)
	if err != nil {
		t.Fatal(err)
	}

	sc := base.NewSlotChain()
	sc.AddRuleCheckSlot(&circuitbreaker.Slot{})
	sc.AddStatSlot(&circuitbreaker.MetricStatSlot{})

	stateListener.On("OnTransformToOpen", circuitbreaker.Closed, mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToHalfOpen", mock.Anything, mock.Anything).Return()

	// First trigger the circuit breaker
	e, b := sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	assert.True(t, b == nil)
	sentinel.TraceError(e, errors.New("biz error"))
	util.Sleep(time.Duration(50) * time.Millisecond)
	e.Exit()
	stateListener.AssertNumberOfCalls(t, "OnTransformToOpen", 2)
	stateListener.AssertNotCalled(t, "OnTransformToClosed")
	stateListener.AssertNotCalled(t, "OnTransformToHalfOpen")

	// wait circuit breaker1 retry timeout
	util.Sleep(time.Duration(100) * time.Millisecond)

	// Second circuit breaker1 probes and circuit breaker2 block the request
	circuitbreaker.ClearStateChangeListeners()
	stateListener2 := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener2)
	stateListener2.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything).Return()
	_, b = sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	assert.True(t, b != nil && b.BlockType() == base.BlockTypeCircuitBreaking && b.TriggeredRule().(*circuitbreaker.Rule) == cbRule2)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToHalfOpen", 1)
	stateListener2.AssertCalled(t, "OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToOpen", 1)
	stateListener2.AssertCalled(t, "OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything)
	util.Sleep(time.Duration(100) * time.Millisecond)

	// Third, same with second request.
	circuitbreaker.ClearStateChangeListeners()
	stateListener3 := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener3)
	stateListener3.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener3.On("OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything).Return()
	stateListener3.On("OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything).Return()
	_, b = sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	assert.True(t, b != nil && b.BlockType() == base.BlockTypeCircuitBreaking && b.TriggeredRule().(*circuitbreaker.Rule) == cbRule2)
	stateListener3.AssertNumberOfCalls(t, "OnTransformToHalfOpen", 1)
	stateListener3.AssertCalled(t, "OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything)
	stateListener3.AssertNumberOfCalls(t, "OnTransformToOpen", 1)
	stateListener3.AssertCalled(t, "OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything)

	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}
}

func TestCircuitBreakerSlotIntegration_Probe_Succeed(t *testing.T) {
	util.SetClock(util.NewMockClock())

	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}

	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		t.Fatal(err)
	}

	cbRule1 := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   20,
		MinRequestAmount: 0,
		StatIntervalMs:   10000,
		MaxAllowedRtMs:   3,
		Threshold:        0.1,
	}

	_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{cbRule1})
	stateListener := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener)
	if err != nil {
		t.Fatal(err)
	}

	sc := base.NewSlotChain()
	sc.AddRuleCheckSlot(&circuitbreaker.Slot{})
	sc.AddStatSlot(&circuitbreaker.MetricStatSlot{})

	stateListener.On("OnTransformToOpen", circuitbreaker.Closed, mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToHalfOpen", mock.Anything, mock.Anything).Return()

	// First trigger the circuit breaker
	e, b := sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	assert.True(t, b == nil)
	util.Sleep(time.Duration(50) * time.Millisecond)
	e.Exit()
	stateListener.AssertNumberOfCalls(t, "OnTransformToOpen", 1)
	stateListener.AssertNotCalled(t, "OnTransformToClosed")
	stateListener.AssertNotCalled(t, "OnTransformToHalfOpen")

	// wait circuit breaker1 retry timeout
	util.Sleep(time.Duration(100) * time.Millisecond)

	// Second circuit breaker1 probes succeed
	circuitbreaker.ClearStateChangeListeners()
	stateListener2 := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener2)
	stateListener2.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything).Return()
	e, b = sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	e.Exit()
	assert.True(t, b == nil)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToHalfOpen", 1)
	stateListener2.AssertCalled(t, "OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToClosed", 1)
	stateListener2.AssertCalled(t, "OnTransformToClosed", circuitbreaker.HalfOpen, mock.Anything)

	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}
}

func TestCircuitBreakerSlotIntegration_Concurrency(t *testing.T) {
	util.SetClock(util.NewMockClock())

	logging.ResetGlobalLoggerLevel(logging.InfoLevel)
	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}
	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		t.Fatal(err)
	}

	cbRule1 := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   1,
		MinRequestAmount: 0,
		StatIntervalMs:   10000,
		MaxAllowedRtMs:   3,
		Threshold:        0.1,
	}
	cbRule2 := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.ErrorRatio,
		RetryTimeoutMs:   2000000,
		MinRequestAmount: 0,
		StatIntervalMs:   10000,
		Threshold:        0.1,
	}

	_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{cbRule1, cbRule2})
	stateListener := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener)
	if err != nil {
		t.Fatal(err)
	}

	sc := base.NewSlotChain()
	sc.AddRuleCheckSlot(&circuitbreaker.Slot{})
	sc.AddStatSlot(&circuitbreaker.MetricStatSlot{})

	stateListener.On("OnTransformToOpen", circuitbreaker.Closed, mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener.On("OnTransformToHalfOpen", mock.Anything, mock.Anything).Return()

	wg := &sync.WaitGroup{}
	wg.Add(100)

	// First trigger the circuit breaker1 and circuit breaker2
	e, b := sentinel.Entry("abc", sentinel.WithSlotChain(sc))
	assert.True(t, b == nil)
	sentinel.TraceError(e, errors.New("biz error"))
	util.Sleep(time.Duration(50) * time.Millisecond)
	e.Exit()
	stateListener.AssertNumberOfCalls(t, "OnTransformToOpen", 2)
	stateListener.AssertNotCalled(t, "OnTransformToClosed")
	stateListener.AssertNotCalled(t, "OnTransformToHalfOpen")
	// wait circuit breaker1 retry timeout
	util.Sleep(time.Duration(100) * time.Millisecond)

	circuitbreaker.ClearStateChangeListeners()
	stateListener2 := &StateChangeListenerMock{}
	circuitbreaker.RegisterStateChangeListeners(stateListener2)
	stateListener2.On("OnTransformToClosed", mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything).Return()
	stateListener2.On("OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything).Return()

	probeFailedCount := int64(0)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				e, b := sentinel.Entry("abc", sentinel.WithSlotChain(sc))
				assert.True(t, b != nil)
				if reflect.DeepEqual(b.TriggeredRule(), cbRule2) {
					atomic.AddInt64(&probeFailedCount, 1)
				}
				if b == nil {
					e.Exit()
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	stateListener2.AssertCalled(t, "OnTransformToHalfOpen", circuitbreaker.Open, mock.Anything)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToHalfOpen", int(atomic.LoadInt64(&probeFailedCount)))
	stateListener2.AssertCalled(t, "OnTransformToOpen", circuitbreaker.HalfOpen, mock.Anything, mock.Anything)
	stateListener2.AssertNumberOfCalls(t, "OnTransformToOpen", int(atomic.LoadInt64(&probeFailedCount)))

	fmt.Println("slow rt rule probe failed: ", atomic.LoadInt64(&probeFailedCount))
	circuitbreaker.ClearStateChangeListeners()
	if clearErr := circuitbreaker.ClearRules(); clearErr != nil {
		t.Fatal(clearErr)
	}
}
