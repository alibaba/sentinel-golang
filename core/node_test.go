package core

import "github.com/stretchr/testify/mock"

type NodeMock struct {
	mock.Mock
}

func (m *NodeMock) TotalCountInMinute() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) PassCountInMinute() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) BlockCountInMinute() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) CompleteCountInMinute() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) ExceptionCountInMinute() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}

func (m *NodeMock) TotalQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) PassQps() uint64 {
	args := m.Called()
	return uint64(*args.Get(0).(*int))
}
func (m *NodeMock) BlockQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) CompleteQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) ExceptionQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) AvgRt() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) CurrentGoroutineNum() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
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
func (m *NodeMock) AddExceptionRequest(count uint64) {
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
