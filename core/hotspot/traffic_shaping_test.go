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

package hotspot

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type counterCacheMock struct {
	mock.Mock
}

func (c *counterCacheMock) Add(key interface{}, value *int64) {
	c.Called(key, value)
	return
}

func (c *counterCacheMock) AddIfAbsent(key interface{}, value *int64) (priorValue *int64) {
	arg := c.Called(key, value)
	ret := arg.Get(0)
	if ret == nil {
		return nil
	}
	return ret.(*int64)
}

func (c *counterCacheMock) Get(key interface{}) (value *int64, isFound bool) {
	arg := c.Called(key)
	val := arg.Get(0)
	if val == nil {
		return nil, arg.Bool(1)
	}
	return val.(*int64), arg.Bool(1)
}

func (c *counterCacheMock) Remove(key interface{}) (isFound bool) {
	arg := c.Called(key)
	return arg.Bool(0)
}

func (c *counterCacheMock) Contains(key interface{}) (ok bool) {
	arg := c.Called(key)
	return arg.Bool(0)
}

func (c *counterCacheMock) Keys() []interface{} {
	arg := c.Called()
	return arg.Get(0).([]interface{})
}

func (c *counterCacheMock) Len() int {
	arg := c.Called()
	return arg.Int(0)
}

func (c *counterCacheMock) Purge() {
	_ = c.Called()
	return
}

func Test_baseTrafficShapingController_performCheckingForConcurrencyMetric(t *testing.T) {
	t.Run("Test_baseTrafficShapingController_performCheckingForConcurrencyMetric", func(t *testing.T) {
		goCounter := &counterCacheMock{}
		c := &baseTrafficShapingController{
			r:             nil,
			res:           "res_a",
			metricType:    Concurrency,
			paramIndex:    0,
			threshold:     100.0,
			specificItems: make(map[interface{}]int64),
			durationInSec: 1,
			metric: &ParamsMetric{
				RuleTimeCounter:    nil,
				RuleTokenCounter:   nil,
				ConcurrencyCounter: goCounter,
			},
		}
		initConcurrency := new(int64)
		*initConcurrency = 50

		goCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(initConcurrency)
		result := c.performCheckingForConcurrencyMetric(666688)
		assert.True(t, result == nil)

		*initConcurrency = 101
		result = c.performCheckingForConcurrencyMetric(666688)
		assert.True(t, result.IsBlocked())

		c.specificItems[666688] = 20
		result = c.performCheckingForConcurrencyMetric(666688)
		assert.True(t, result.IsBlocked())
	})
}

func Test_defaultTrafficShapingController_performChecking(t *testing.T) {
	util.SetClock(util.NewMockClock())

	t.Run("Test_defaultTrafficShapingController_performChecking_TimeCounter_Nil", func(t *testing.T) {
		timeCounter := &counterCacheMock{}
		tokenCounter := &counterCacheMock{}
		goCounter := &counterCacheMock{}
		c := &rejectTrafficShapingController{
			baseTrafficShapingController: baseTrafficShapingController{
				r:             nil,
				res:           "res_a",
				metricType:    QPS,
				paramIndex:    0,
				threshold:     100.0,
				specificItems: make(map[interface{}]int64),
				durationInSec: 1,
				metric: &ParamsMetric{
					RuleTimeCounter:    timeCounter,
					RuleTokenCounter:   tokenCounter,
					ConcurrencyCounter: goCounter,
				},
			},
			burstCount: 10,
		}
		arg := 010110
		result := c.PerformChecking(arg, 130)
		assert.True(t, result.IsBlocked())

		lastAddTokenTime := new(int64)
		*lastAddTokenTime = 1578416556900
		timeCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Times(1).Return(nil)
		tokenCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Times(1).Return(nil)
		result = c.PerformChecking(arg, 20)
		assert.True(t, result == nil)
	})

	t.Run("Test_defaultTrafficShapingController_performChecking_Sub_Token", func(t *testing.T) {
		timeCounter := &counterCacheMock{}
		tokenCounter := &counterCacheMock{}
		c := &rejectTrafficShapingController{
			baseTrafficShapingController: baseTrafficShapingController{
				r:             nil,
				res:           "res_a",
				metricType:    QPS,
				paramIndex:    0,
				threshold:     100.0,
				specificItems: make(map[interface{}]int64),
				durationInSec: 10,
				metric: &ParamsMetric{
					RuleTimeCounter:    timeCounter,
					RuleTokenCounter:   tokenCounter,
					ConcurrencyCounter: nil,
				},
			},
			burstCount: 10,
		}
		arg := 010110
		lastAddTokenTime := new(int64)
		currentTimeInMs := int64(util.CurrentTimeMillis())
		*lastAddTokenTime = currentTimeInMs - 1000
		timeCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Times(1).Return(lastAddTokenTime)
		oldQps := new(int64)
		*oldQps = 50
		tokenCounter.On("Get", mock.Anything).Return(oldQps, true).Times(1)
		result := c.PerformChecking(arg, 20)
		assert.True(t, result == nil)
		assert.True(t, atomic.LoadInt64(oldQps) == 30)
	})

	t.Run("Test_defaultTrafficShapingController_performChecking_First_Fill_Token", func(t *testing.T) {
		timeCounter := &counterCacheMock{}
		tokenCounter := &counterCacheMock{}
		c := &rejectTrafficShapingController{
			baseTrafficShapingController: baseTrafficShapingController{
				r:             nil,
				res:           "res_a",
				metricType:    QPS,
				paramIndex:    0,
				threshold:     100.0,
				specificItems: make(map[interface{}]int64),
				durationInSec: 1,
				metric: &ParamsMetric{
					RuleTimeCounter:    timeCounter,
					RuleTokenCounter:   tokenCounter,
					ConcurrencyCounter: nil,
				},
			},
			burstCount: 10,
		}
		arg := 010110
		lastAddTokenTime := new(int64)
		currentTimeInMs := int64(util.CurrentTimeMillis())
		*lastAddTokenTime = currentTimeInMs - 1001
		timeCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(lastAddTokenTime).Times(1)

		tokenCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(nil).Times(1)
		util.Sleep(time.Duration(10) * time.Millisecond)
		result := c.PerformChecking(arg, 20)
		assert.True(t, result == nil)
		assert.True(t, *lastAddTokenTime > currentTimeInMs)
	})

	t.Run("Test_defaultTrafficShapingController_performChecking_Refill_Token", func(t *testing.T) {
		timeCounter := &counterCacheMock{}
		tokenCounter := &counterCacheMock{}
		c := &rejectTrafficShapingController{
			baseTrafficShapingController: baseTrafficShapingController{
				r:             nil,
				res:           "res_a",
				metricType:    QPS,
				paramIndex:    0,
				threshold:     100.0,
				specificItems: make(map[interface{}]int64),
				durationInSec: 1,
				metric: &ParamsMetric{
					RuleTimeCounter:    timeCounter,
					RuleTokenCounter:   tokenCounter,
					ConcurrencyCounter: nil,
				},
			},
			burstCount: 10,
		}
		arg := 010110
		lastAddTokenTime := new(int64)
		currentTimeInMs := int64(util.CurrentTimeMillis())
		*lastAddTokenTime = currentTimeInMs - 1001
		timeCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(lastAddTokenTime).Times(1)

		oldQps := new(int64)
		*oldQps = 50
		tokenCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(oldQps).Times(1)
		util.Sleep(time.Duration(10) * time.Millisecond)
		result := c.PerformChecking(arg, 20)
		assert.True(t, result == nil)
		assert.True(t, atomic.LoadInt64(lastAddTokenTime) > currentTimeInMs)
		assert.True(t, atomic.LoadInt64(oldQps) > 30)
	})
}

func Test_throttlingTrafficShapingController_performChecking(t *testing.T) {
	t.Run("Test_throttlingTrafficShapingController_performChecking", func(t *testing.T) {
		timeCounter := &counterCacheMock{}
		tokenCounter := &counterCacheMock{}
		c := &throttlingTrafficShapingController{
			baseTrafficShapingController: baseTrafficShapingController{
				r:             nil,
				res:           "res_a",
				metricType:    QPS,
				paramIndex:    0,
				threshold:     100.0,
				specificItems: make(map[interface{}]int64),
				durationInSec: 1,
				metric: &ParamsMetric{
					RuleTimeCounter:    timeCounter,
					RuleTokenCounter:   tokenCounter,
					ConcurrencyCounter: nil,
				},
			},
			maxQueueingTimeMs: 10,
		}

		arg := 010110
		lastAddTokenTime := new(int64)
		currentTimeInMs := int64(util.CurrentTimeMillis())
		*lastAddTokenTime = currentTimeInMs - 201
		timeCounter.On("AddIfAbsent", mock.Anything, mock.Anything).Return(lastAddTokenTime).Times(1)
		result := c.PerformChecking(arg, 20)
		assert.True(t, result == nil)
	})
}

func Test_newBaseTrafficShapingController(t *testing.T) {
	t.Run("Test_newBaseTrafficShapingController", func(t *testing.T) {
		tc := newBaseTrafficShapingController(&Rule{
			MetricType: Concurrency,
		})
		for i := 0; i < 10000; i++ {
			initConcurrency := new(int64)
			*initConcurrency = 0
			tc.metric.ConcurrencyCounter.AddIfAbsent(i, initConcurrency)
		}
		assert.True(t, tc.metric.ConcurrencyCounter != nil)
		assert.True(t, tc.metric.ConcurrencyCounter.Len() == ConcurrencyMaxCount)

		tc = newBaseTrafficShapingController(&Rule{
			MetricType:        Concurrency,
			ParamsMaxCapacity: 100,
		})
		for i := 0; i < 10000; i++ {
			initConcurrency := new(int64)
			*initConcurrency = 0
			tc.metric.ConcurrencyCounter.AddIfAbsent(i, initConcurrency)
		}
		assert.True(t, tc.metric.ConcurrencyCounter != nil)
		assert.True(t, tc.metric.ConcurrencyCounter.Len() == 100)

		tc = newBaseTrafficShapingController(&Rule{
			MetricType: QPS,
		})
		for i := 0; i < 30000; i++ {
			initConcurrency := new(int64)
			*initConcurrency = 0
			tc.metric.RuleTimeCounter.AddIfAbsent(i, initConcurrency)
			tc.metric.RuleTokenCounter.AddIfAbsent(i, initConcurrency)
		}
		assert.True(t, tc.metric.RuleTimeCounter != nil)
		assert.True(t, tc.metric.RuleTokenCounter != nil)
		assert.True(t, tc.metric.RuleTimeCounter.Len() == ParamsMaxCapacity)
		assert.True(t, tc.metric.RuleTokenCounter.Len() == ParamsMaxCapacity)
	})
}

func Test_baseTrafficShapingController_ExtractArgs(t *testing.T) {
	t.Run("Test_baseTrafficShapingController_ExtractArgs", func(t *testing.T) {

		c := &baseTrafficShapingController{}

		args := make([]interface{}, 10)
		attachments := make(map[interface{}]interface{})
		ctx := base.NewEmptyEntryContext()
		ctx.Input = &base.SentinelInput{
			BatchCount:  1,
			Flag:        0,
			Args:        args,
			Attachments: attachments,
		}
		// no data
		ret := c.ExtractArgs(ctx)
		assert.Nil(t, ret)

		// set data
		args[0] = 1
		args[1] = 2
		value1 := "v1"
		attachments["test1"] = value1

		// set index or key
		// exist
		c.paramIndex = 0
		c.paramKey = "test1"
		ret = c.ExtractArgs(ctx)
		assert.True(t, reflect.DeepEqual(ret, value1), ret)

		// part exist 1
		c.paramIndex = 10
		c.paramKey = "test1"
		ret = c.ExtractArgs(ctx)
		assert.True(t, reflect.DeepEqual(ret, value1), ret)

		// part exist 2
		c.paramIndex = 1
		c.paramKey = "test2"
		ret = c.ExtractArgs(ctx)
		assert.True(t, reflect.DeepEqual(ret, 2), ret)

		// not exist
		c.paramIndex = 3
		c.paramKey = "test2"
		ret = c.ExtractArgs(ctx)
		assert.Nil(t, ret)
	})
}
