package base

import "github.com/stretchr/testify/mock"

type StatNodeMock struct {
	mock.Mock
}

func (m *StatNodeMock) MetricsOnCondition(predicate TimePredicate) []*MetricItem {
	args := m.Called()
	return args.Get(0).([]*MetricItem)
}

func (m *StatNodeMock) GetQPS(event MetricEvent) float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) GetQPSWithTime(now uint64, event MetricEvent) float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) TotalQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) GetSum(event MetricEvent) int64 {
	args := m.Called()
	return int64(args.Int(0))
}

func (m *StatNodeMock) GetSumWithTime(now uint64, event MetricEvent) int64 {
	args := m.Called()
	return int64(args.Int(0))
}

func (m *StatNodeMock) AvgRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *StatNodeMock) MinRT() int64 {
	args := m.Called()
	return int64(args.Int(0))
}

func (m *StatNodeMock) AddRequest(event MetricEvent, count uint64) {
	m.Called(event, count)
	return
}

func (m *StatNodeMock) AddRtAndCompleteRequest(rt, count uint64) {
	m.Called(rt, count)
	return
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
