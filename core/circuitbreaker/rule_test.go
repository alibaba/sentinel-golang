package circuitbreaker

import (
	"github.com/stretchr/testify/mock"
)

type RuleMock struct {
	mock.Mock
}

func (m *RuleMock) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *RuleMock) ResourceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *RuleMock) BreakerStrategy() Strategy {
	args := m.Called()
	return args.Get(0).(Strategy)
}
func (m *RuleMock) BreakerStatIntervalMs() uint32 {
	args := m.Called()
	return uint32(args.Int(0))
}

func (m *RuleMock) IsEqualsTo(r Rule) bool {
	args := m.Called(r)
	return args.Bool(0)
}

func (m *RuleMock) IsStatReusable(r Rule) bool {
	args := m.Called(r)
	return args.Bool(0)
}

func (m *RuleMock) IsApplicable() error {
	args := m.Called()
	return args.Get(0).(error)
}
