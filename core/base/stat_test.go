package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StatNodeMock struct {
	mock.Mock
}

func (m *StatNodeMock) AddCount(event MetricEvent, count int64) {
	m.Called(event, count)
}

func (m *StatNodeMock) MetricsOnCondition(predicate TimePredicate) []*MetricItem {
	args := m.Called(predicate)
	return args.Get(0).([]*MetricItem)
}

func (m *StatNodeMock) GetQPS(event MetricEvent) float64 {
	args := m.Called(event)
	return float64(args.Int(0))
}

func (m *StatNodeMock) GetPreviousQPS(event MetricEvent) float64 {
	args := m.Called(event)
	return args.Get(0).(float64)
}

func (m *StatNodeMock) GetMaxAvg(event MetricEvent) float64 {
	args := m.Called(event)
	return float64(args.Int(0))
}

func (m *StatNodeMock) GetSum(event MetricEvent) int64 {
	args := m.Called(event)
	return int64(args.Int(0))
}

func (m *StatNodeMock) AvgRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) MinRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) CurrentGoroutineNum() int32 {
	args := m.Called()
	return int32(args.Int(0))
}

func (m *StatNodeMock) IncreaseGoroutineNum() {
	m.Called()
	return
}

func (m *StatNodeMock) DecreaseGoroutineNum() {
	m.Called()
	return
}

func (m *StatNodeMock) Reset() {
	m.Called()
	return
}

func (m *StatNodeMock) GenerateReadStat(sampleCount uint32, intervalInMs uint32) (ReadStat, error) {
	args := m.Called(sampleCount, intervalInMs)
	return args.Get(0).(ReadStat), args.Error(1)
}

func TestCheckValidityForReuseStatistic(t *testing.T) {
	assert.Equal(t, CheckValidityForReuseStatistic(3, 1000, 20, 10000), IllegalStatisticParamsError)
	assert.Equal(t, CheckValidityForReuseStatistic(0, 1000, 20, 10000), IllegalStatisticParamsError)
	assert.Equal(t, CheckValidityForReuseStatistic(2, 1000, 21, 10000), IllegalGlobalStatisticParamsError)
	assert.Equal(t, CheckValidityForReuseStatistic(2, 1000, 0, 10000), IllegalGlobalStatisticParamsError)
	assert.Equal(t, CheckValidityForReuseStatistic(2, 8000, 20, 10000), GlobalStatisticNonReusableError)
	assert.Equal(t, CheckValidityForReuseStatistic(2, 1000, 10, 10000), GlobalStatisticNonReusableError)
	assert.Equal(t, CheckValidityForReuseStatistic(1, 1000, 100, 10000), nil)
	assert.Equal(t, CheckValidityForReuseStatistic(2, 1000, 20, 10000), nil)
}
