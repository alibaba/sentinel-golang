package base

import "github.com/stretchr/testify/mock"

type StatNodeMock struct {
	mock.Mock
}

func (m *StatNodeMock) AddMetric(event MetricEvent, count uint64) {
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
