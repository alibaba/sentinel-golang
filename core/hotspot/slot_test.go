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
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TrafficShapingControllerMock struct {
	mock.Mock
}

func (m *TrafficShapingControllerMock) PerformChecking(arg interface{}, batchCount int64) *base.TokenResult {
	retArgs := m.Called(arg, batchCount)
	return retArgs.Get(0).(*base.TokenResult)
}

func (m *TrafficShapingControllerMock) BoundParamIndex() int {
	retArgs := m.Called()
	return retArgs.Int(0)
}

func (m *TrafficShapingControllerMock) BoundMetric() *ParamsMetric {
	retArgs := m.Called()
	return retArgs.Get(0).(*ParamsMetric)
}

func (m *TrafficShapingControllerMock) BoundRule() *Rule {
	retArgs := m.Called()
	return retArgs.Get(0).(*Rule)
}

func (m *TrafficShapingControllerMock) Replace(r *Rule) {
	_ = m.Called(r)
	return
}

func Test_matchArg(t *testing.T) {
	t.Run("Test_matchArg", func(t *testing.T) {
		args := make([]interface{}, 10)
		args[0] = 1
		args[1] = 2
		args[2] = 3
		args[3] = 4
		args[4] = 5
		args[5] = 6
		args[6] = 7
		args[7] = 8
		args[8] = 9
		args[9] = 10

		tcMock := &TrafficShapingControllerMock{}
		tcMock.On("BoundParamIndex").Return(0)
		ret0 := matchArg(tcMock, args)
		assert.True(t, reflect.DeepEqual(ret0, 1))

		tcMock1 := &TrafficShapingControllerMock{}
		tcMock1.On("BoundParamIndex").Return(5)
		ret1 := matchArg(tcMock1, args)
		assert.True(t, reflect.DeepEqual(ret1, 6))

		tcMock2 := &TrafficShapingControllerMock{}
		tcMock2.On("BoundParamIndex").Return(9)
		ret2 := matchArg(tcMock2, args)
		assert.True(t, reflect.DeepEqual(ret2, 10))

		tcMock3 := &TrafficShapingControllerMock{}
		tcMock3.On("BoundParamIndex").Return(-1)
		ret3 := matchArg(tcMock3, args)
		assert.True(t, reflect.DeepEqual(ret3, 10))

		tcMock4 := &TrafficShapingControllerMock{}
		tcMock4.On("BoundParamIndex").Return(-10)
		ret4 := matchArg(tcMock4, args)
		assert.True(t, reflect.DeepEqual(ret4, 1))

		tcMock5 := &TrafficShapingControllerMock{}
		tcMock5.On("BoundParamIndex").Return(10)
		ret5 := matchArg(tcMock5, args)
		assert.True(t, ret5 == nil)

		tcMock6 := &TrafficShapingControllerMock{}
		tcMock6.On("BoundParamIndex").Return(-11)
		ret6 := matchArg(tcMock6, args)
		assert.True(t, ret6 == nil)
	})
}
