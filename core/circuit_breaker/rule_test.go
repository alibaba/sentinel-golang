package circuit_breaker

import "github.com/stretchr/testify/mock"

type ruleMock struct {
	mock.Mock
}

func (m *ruleMock) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *ruleMock) ResourceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *ruleMock) BreakerStrategy() BreakerStrategy {
	args := m.Called()
	return args.Get(0).(BreakerStrategy)
}

func (m *ruleMock) isApplicable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *ruleMock) convert2CircuitBreaker() CircuitBreaker {
	args := m.Called()
	return args.Get(0).(CircuitBreaker)
}
