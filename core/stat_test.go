package core

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

func (m *NodeMock) PassQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) BlockQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) CompleteQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *NodeMock) ErrorQPS() float64 {
	args := m.Called()
	return float64(args.Int(0))
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

func (m *NodeMock) AddPassRequest(count uint64) {
	m.Called(count)
	return
}

func (m *NodeMock) AddRtAndCompleteRequest(rt, count uint64) {
	m.Called(rt, count)
	return
}

func (m *NodeMock) AddBlockRequest(count uint64) {
	m.Called(count)
	return
}

func (m *NodeMock) AddErrorRequest(count uint64) {
	m.Called(count)
	return
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
