package base

import "github.com/stretchr/testify/mock"

type NodeMock struct {
	mock.Mock
}

func (m *NodeMock) MetricsOnCondition(predicate TimePredicate) []*MetricItem {
	args := m.Called()
	return args.Get(0).([]*MetricItem)
}

func (m *NodeMock) TotalQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) GetQPS(event MetricEvent) float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) AddRequest(event MetricEvent, count uint64) {
	m.Called(event,count)
	return
}

func (m *NodeMock) AddRtAndCompleteRequest(rt, count uint64) {
	m.Called(rt, count)
	return
}

func (m *NodeMock) AvgRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) MinRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) CurrentGoroutineNum() int32 {
	args := m.Called()
	return int32(args.Int(0))
}

func (m *NodeMock) IncreaseGoroutineNum() {
	m.Called()
	return
}

func (m *NodeMock) DecreaseGoroutineNum() {
	m.Called()
	return
}

func (m *NodeMock) Reset() {
	m.Called()
	return
}
