package stat

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/base/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatisticSlot_String(t *testing.T) {
	s := &StatisticSlot{}

	assert.Equal(t, SlotName, s.String())
}

func TestStatisticSlot_OnEntryPassed(t *testing.T) {
	s := &StatisticSlot{}
	sInput := &base.SentinelInput{
		AcquireCount: 1,
	}
	m := mocks.StatNode{}
	m.On("IncreaseGoroutineNum").Return()
	m.On("AddMetric", mock.Anything, mock.Anything).Return()

	t.Run("NonInboundFlowType", func(t *testing.T) {
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Outbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
		}
		s.OnEntryPassed(e)
	})

	t.Run("InboundFlowType", func(t *testing.T) {
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Inbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
		}
		s.OnEntryPassed(e)
	})
}

func TestStatisticSlot_OnEntryBlocked(t *testing.T) {
	s := &StatisticSlot{}
	sInput := &base.SentinelInput{
		AcquireCount: 1,
	}
	m := mocks.StatNode{}
	m.On("IncreaseGoroutineNum").Return()
	m.On("AddMetric", mock.Anything, mock.Anything).Return()

	t.Run("NonInboundFlowType", func(t *testing.T) {
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Outbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
		}
		s.OnEntryBlocked(e, nil)
	})

	t.Run("InboundFlowType", func(t *testing.T) {
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Inbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
		}
		s.OnEntryBlocked(e, nil)
	})
}

func TestStatisticSlot_OnCompleted(t *testing.T) {
	s := &StatisticSlot{}

	t.Run("NilLastResult", func(t *testing.T) {
		sOutput := &base.SentinelOutput{}
		e := &base.EntryContext{
			Output: sOutput,
		}
		s.OnCompleted(e)
	})

	t.Run("TrueIsBlocked", func(t *testing.T) {
		tResult := base.NewTokenResultBlocked(base.BlockTypeFlow, "TrueIsBlocked")
		sOutput := &base.SentinelOutput{
			LastResult: tResult,
		}
		e := &base.EntryContext{
			Output: sOutput,
		}
		s.OnCompleted(e)
	})

	t.Run("NormalNonInboundFlowType", func(t *testing.T) {
		m := mocks.StatNode{}
		m.On("DecreaseGoroutineNum").Return()
		m.On("AddMetric", mock.Anything, mock.Anything).Return()
		tResult := base.NewTokenResultPass()
		sOutput := &base.SentinelOutput{
			LastResult: tResult,
		}
		sInput := &base.SentinelInput{
			AcquireCount: 1,
		}
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Outbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
			Output:   sOutput,
		}
		s.OnCompleted(e)
	})

	t.Run("NormalInboundFlowType", func(t *testing.T) {
		m := mocks.StatNode{}
		m.On("DecreaseGoroutineNum").Return()
		m.On("AddMetric", mock.Anything, mock.Anything).Return()
		tResult := base.NewTokenResultPass()
		sOutput := &base.SentinelOutput{
			LastResult: tResult,
		}
		sInput := &base.SentinelInput{
			AcquireCount: 1,
		}
		r := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Inbound)
		e := &base.EntryContext{
			Resource: r,
			StatNode: &m,
			Input:    sInput,
			Output:   sOutput,
		}
		s.OnCompleted(e)
	})
}

func TestStatisticSlot_recordPassFor(t *testing.T) {
	t.Run("NilStatNode", func(t *testing.T) {
		s := &StatisticSlot{}

		s.recordPassFor(nil, 0)
	})

	t.Run("NormalStatNode", func(t *testing.T) {
		s := &StatisticSlot{}
		m := mocks.StatNode{}
		m.On("IncreaseGoroutineNum").Return()
		m.On("AddMetric", base.MetricEventPass, uint64(1)).Return()
		s.recordPassFor(&m, 1)
	})
}

func TestStatisticSlot_recordBlockFor(t *testing.T) {
	t.Run("NilStatNode", func(t *testing.T) {
		s := &StatisticSlot{}

		s.recordBlockFor(nil, 0)
	})

	t.Run("NormalStatNode", func(t *testing.T) {
		s := &StatisticSlot{}
		m := mocks.StatNode{}
		m.On("AddMetric", base.MetricEventBlock, uint64(1)).Return()
		s.recordBlockFor(&m, 1)
	})
}

func TestStatisticSlot_recordCompleteFor(t *testing.T) {
	t.Run("NilStatNode", func(t *testing.T) {
		s := &StatisticSlot{}

		s.recordCompleteFor(nil, 0, 0)
	})

	t.Run("NormalStatNode", func(t *testing.T) {
		s := &StatisticSlot{}
		m := mocks.StatNode{}
		m.On("DecreaseGoroutineNum").Return()
		m.On("AddMetric", mock.Anything, mock.Anything).Return()
		s.recordCompleteFor(&m, 1, 1)
	})
}
