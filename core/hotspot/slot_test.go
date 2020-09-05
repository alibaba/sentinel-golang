package hotspot

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/stretchr/testify/mock"
)

type TrafficShapingControllerMock struct {
	mock.Mock
}

func (m *TrafficShapingControllerMock) PerformChecking(arg interface{}, acquireCount int64) *base.TokenResult {
	retArgs := m.Called(arg, acquireCount)
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
		args[0] = true
		args[1] = false
		args[2] = float32(1.2345678)
		args[3] = float64(1.23)
		args[4] = uint8(66)
		args[5] = int32(88)
		args[6] = int(6688)
		args[7] = uint64(668866)
		args[8] = "ximu"
		args[9] = int64(-100)

		tcMock := &TrafficShapingControllerMock{}
		tcMock.On("BoundParamIndex").Return(0)
		ret0 := matchArg(tcMock, args)
		assert.True(t, reflect.TypeOf(ret0).Kind() == reflect.Bool && ret0 == true)

		tcMock1 := &TrafficShapingControllerMock{}
		tcMock1.On("BoundParamIndex").Return(1)
		ret1 := matchArg(tcMock1, args)
		assert.True(t, reflect.TypeOf(ret1).Kind() == reflect.Bool && ret1 == false)

		tcMock2 := &TrafficShapingControllerMock{}
		tcMock2.On("BoundParamIndex").Return(2)
		ret2 := matchArg(tcMock2, args)
		assert.True(t, reflect.TypeOf(ret2).Kind() == reflect.Float64 && ret2 == 1.23457)

		tcMock3 := &TrafficShapingControllerMock{}
		tcMock3.On("BoundParamIndex").Return(3)
		ret3 := matchArg(tcMock3, args)
		assert.True(t, reflect.TypeOf(ret3).Kind() == reflect.Float64 && ret3 == 1.23000)

		tcMock4 := &TrafficShapingControllerMock{}
		tcMock4.On("BoundParamIndex").Return(4)
		ret4 := matchArg(tcMock4, args)
		assert.True(t, reflect.TypeOf(ret4).Kind() == reflect.Int && ret4 == 66)

		tcMock5 := &TrafficShapingControllerMock{}
		tcMock5.On("BoundParamIndex").Return(5)
		ret5 := matchArg(tcMock5, args)
		assert.True(t, reflect.TypeOf(ret5).Kind() == reflect.Int && ret5 == 88)

		tcMock6 := &TrafficShapingControllerMock{}
		tcMock6.On("BoundParamIndex").Return(6)
		ret6 := matchArg(tcMock6, args)
		assert.True(t, reflect.TypeOf(ret6).Kind() == reflect.Int && ret6 == 6688)

		tcMock7 := &TrafficShapingControllerMock{}
		tcMock7.On("BoundParamIndex").Return(7)
		ret7 := matchArg(tcMock7, args)
		assert.True(t, reflect.TypeOf(ret7).Kind() == reflect.Int && ret7 == 668866)

		tcMock8 := &TrafficShapingControllerMock{}
		tcMock8.On("BoundParamIndex").Return(8)
		ret8 := matchArg(tcMock8, args)
		assert.True(t, reflect.TypeOf(ret8).Kind() == reflect.String && ret8 == "ximu")

		tcMock9 := &TrafficShapingControllerMock{}
		tcMock9.On("BoundParamIndex").Return(9)
		ret9 := matchArg(tcMock9, args)
		assert.True(t, reflect.TypeOf(ret9).Kind() == reflect.Int && ret9 == -100)

	})
}
