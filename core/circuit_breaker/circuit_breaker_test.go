package circuit_breaker

import (
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/mock"
)

type ReadStatMock struct {
	mock.Mock
}

func (m *ReadStatMock) GetQPS(event base.MetricEvent) float64 {
	args := m.Called(event)
	return float64(args.Int(0))
}

func (m *ReadStatMock) GetQPSWithTime(now uint64, event base.MetricEvent) float64 {
	args := m.Called(now, event)
	return float64(args.Int(0))
}

func (m *ReadStatMock) GetSum(event base.MetricEvent) int64 {
	args := m.Called(event)
	return int64(args.Int(0))
}

func (m *ReadStatMock) GetSumWithTime(now uint64, event base.MetricEvent) int64 {
	args := m.Called(now, event)
	return int64(args.Int(0))
}

func (m *ReadStatMock) AvgRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

func (m *ReadStatMock) MinRT() float64 {
	args := m.Called()
	return float64(args.Int(0))
}

type circuitBreakerMock struct {
	mock.Mock
}

func (m *circuitBreakerMock) getRule() Rule {
	args := m.Called()
	return args.Get(0).(Rule)
}

func (m *circuitBreakerMock) Check(_ *base.EntryContext) *base.TokenResult {
	args := m.Called()
	return args.Get(0).(*base.TokenResult)
}

func Test_AverageRtCircuitBreaker_Check(t *testing.T) {
	type args struct {
		ctx *base.EntryContext
	}
	m := &ReadStatMock{}
	rule := &averageRtRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       "abc01",
			Strategy:       AverageRt,
			RecoverTimeout: 1,
			SampleCount:    2,
			IntervalInMs:   1000,
		},
		Threshold:           100,
		RtSlowRequestAmount: 5,
	}
	tests := []struct {
		name    string
		breaker *averageRtCircuitBreaker
		args    args
	}{
		{
			name:    "Test_AverageRtCircuitBreaker_Check",
			breaker: newAverageRtCircuitBreakerWithMetric(rule, m),
			args: args{
				ctx: nil,
			},
		},
	}

	m.On("AvgRT").Return(100)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 4; i++ {
				if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
					t.Errorf("averageRtCircuitBreaker.Check() = %v, want %v", got, true)
				}
			}

			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != false {
				t.Errorf("averageRtCircuitBreaker.Check() = %v, want %v", got, false)
			}

			// before auto recover
			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != false {
				t.Errorf("averageRtCircuitBreaker.Check() = %v, want %v", got, false)
			}

			time.Sleep(2 * time.Second)
			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
				t.Errorf("averageRtCircuitBreaker.Check() = %v, want %v", got, true)
			}
		})
	}
}

func Test_ErrorRatioCircuitBreaker_Check(t *testing.T) {
	type args struct {
		ctx *base.EntryContext
	}
	m := &ReadStatMock{}
	rule := &errorRatioRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       "abc01",
			Strategy:       ErrorRatio,
			RecoverTimeout: 1,
			SampleCount:    2,
			IntervalInMs:   1000,
		},
		Threshold:        0.3,
		MinRequestAmount: 5,
	}
	tests := []struct {
		name    string
		breaker *errorRatioCircuitBreaker
		args    args
	}{
		{
			name:    "Test_ErrorRatioCircuitBreaker_Check",
			breaker: newErrorRatioCircuitBreakerWithMetric(rule, m),
			args: args{
				ctx: nil,
			},
		},
	}

	// mock data
	m.On("GetQPS", base.MetricEventError).Return(100)
	m.On("GetQPS", base.MetricEventComplete).Return(400)
	m.On("GetQPS", base.MetricEventPass).Return(800)
	m.On("GetQPS", base.MetricEventBlock).Return(200)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 4; i++ {
				if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
					t.Errorf("ErrorRatioCircuitBreaker.Check() = %v, want %v", got, true)
				}
			}

			m2 := &ReadStatMock{}
			tt.breaker.metric = m2
			m2.On("GetQPS", base.MetricEventError).Return(200)
			m2.On("GetQPS", base.MetricEventComplete).Return(400)
			m2.On("GetQPS", base.MetricEventPass).Return(800)
			m2.On("GetQPS", base.MetricEventBlock).Return(200)

			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != false {
				t.Errorf("ErrorRatioCircuitBreaker.Check() = %v, want %v", got, false)
			}
			time.Sleep(2 * time.Second)

			m3 := &ReadStatMock{}
			tt.breaker.metric = m3
			m3.On("GetQPS", base.MetricEventError).Return(0)
			m3.On("GetQPS", base.MetricEventComplete).Return(0)

			m3.On("GetQPS", base.MetricEventPass).Return(0)
			m3.On("GetQPS", base.MetricEventBlock).Return(0)
			m3.On("GetQPS", base.MetricEventPass).Return(0)
			m3.On("GetQPS", base.MetricEventBlock).Return(0)
			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
				t.Errorf("ErrorRatioCircuitBreaker.Check() = %v, want %v", got, true)
			}
		})
	}
}

func Test_ErrorCountCircuitBreaker_Check(t *testing.T) {
	type args struct {
		ctx *base.EntryContext
	}
	m := &ReadStatMock{}
	rule := &errorCountRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       "abc01",
			Strategy:       ErrorCount,
			RecoverTimeout: 1,
			SampleCount:    2,
			IntervalInMs:   1000,
		},
		Threshold: 10,
	}

	tests := []struct {
		name    string
		breaker *errorCountCircuitBreaker
		args    args
	}{
		{
			name:    "Test_ErrorCountCircuitBreaker_Check",
			breaker: newErrorCountCircuitBreakerWithMetric(rule, m),
			args: args{
				ctx: nil,
			},
		},
	}
	m.On("GetQPS", base.MetricEventError).Return(5)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 4; i++ {
				if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
					t.Errorf("ErrorCountCircuitBreaker.Check() = %v, want %v", got, true)
				}
			}

			m2 := &ReadStatMock{}
			tt.breaker.metric = m2
			m2.On("GetQPS", base.MetricEventError).Return(11)
			for i := 0; i < 10; i++ {
				if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != false {
					t.Errorf("ErrorCountCircuitBreaker.Check() = %v, want %v", got, true)
				}
			}
			time.Sleep(2 * time.Second)

			m3 := &ReadStatMock{}
			tt.breaker.metric = m3
			m3.On("GetQPS", base.MetricEventError).Return(1)
			if got := tt.breaker.Check(tt.args.ctx); got.IsPass() != true {
				t.Errorf("ErrorCountCircuitBreaker.Check() = %v, want %v", got, true)
			}
		})
	}
}
