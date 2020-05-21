package circuitbreaker

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

type CircuitBreakerMock struct {
	mock.Mock
}

func (m *CircuitBreakerMock) BoundRule() Rule {
	args := m.Called()
	return args.Get(0).(Rule)
}

func (m *CircuitBreakerMock) TryPass(ctx *base.EntryContext) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *CircuitBreakerMock) CurrentState() State {
	args := m.Called()
	return args.Get(0).(State)
}

// OnRequestComplete handle the entry completed
// rt: the response time this entry cost.
func (m *CircuitBreakerMock) HandleCompleted(rt uint64, err error) {
	m.Called(rt, err)
	return
}

func TestStatus(t *testing.T) {
	t.Run("get_set", func(t *testing.T) {
		status := new(State)
		assert.True(t, status.get() == Closed)

		status.set(Open)
		assert.True(t, status.get() == Open)
	})

	t.Run("cas", func(t *testing.T) {
		status := new(State)
		assert.True(t, status.get() == Closed)

		assert.True(t, status.casState(Closed, Open))
		assert.True(t, !status.casState(Closed, Open))
		status.set(HalfOpen)
		assert.True(t, status.casState(HalfOpen, Open))
	})
}
