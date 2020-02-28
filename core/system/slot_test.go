package system

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"

	"github.com/stretchr/testify/assert"
)

func TestCheckNilInput(t *testing.T) {
	var sas *SystemAdaptiveSlot
	e := base.NewTokenResultPass()

	t.Run("NilInput", func(t *testing.T) {
		r := sas.Check(nil)
		assert.Equal(t, e, r)
	})

	t.Run("NilResourceInput", func(t *testing.T) {
		r := sas.Check(&base.EntryContext{})
		assert.Equal(t, e, r)
	})

	t.Run("UnsuitableFlowType", func(t *testing.T) {
		rw := base.NewResourceWrapper("test", base.ResTypeCommon, base.Outbound)
		r := sas.Check(&base.EntryContext{Resource: rw})
		assert.Equal(t, e, r)
	})
}

func TestCheckEmptyRule(t *testing.T) {
	var sas *SystemAdaptiveSlot
	e := base.NewTokenResultPass()
	rw := base.NewResourceWrapper("test", base.ResTypeCommon, base.Inbound)
	r := sas.Check(&base.EntryContext{Resource: rw})
	assert.Equal(t, e, r)
}

func TestDoCheckRuleConcurrency(t *testing.T) {
	var sas *SystemAdaptiveSlot
	rule := &SystemRule{MetricType: Concurrency,
		TriggerCount: 0.5}

	t.Run("TrueConcurrency", func(t *testing.T) {
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, float64(0), v)
		assert.Equal(t, true, isOK)
	})

	t.Run("FalseConcurrency", func(t *testing.T) {
		stat.InboundNode().IncreaseGoroutineNum()
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, float64(1), v)
		assert.Equal(t, false, isOK)
		stat.InboundNode().DecreaseGoroutineNum()
	})
}

func TestDoCheckRuleLoad(t *testing.T) {
	var sas *SystemAdaptiveSlot
	rule := &SystemRule{MetricType: Load,
		TriggerCount: 0.5}

	t.Run("TrueLoad", func(t *testing.T) {
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, notRetrievedValue, v)
		assert.Equal(t, true, isOK)
	})

	t.Run("BBRTrueLoad", func(t *testing.T) {
		rule.Strategy = BBR
		currentLoad.Store(float64(1))
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, true, isOK)
		assert.Equal(t, float64(1), v)
		currentLoad.Store(float64(notRetrievedValue))
	})
}

func TestDoCheckRuleCpuUsage(t *testing.T) {
	var sas *SystemAdaptiveSlot
	rule := &SystemRule{MetricType: CpuUsage,
		TriggerCount: 0.5}

	t.Run("TrueCpuUsage", func(t *testing.T) {
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, notRetrievedValue, v)
		assert.Equal(t, true, isOK)
	})

	t.Run("BBRTrueCpuUsage", func(t *testing.T) {
		rule.Strategy = BBR
		currentCpuUsage.Store(float64(0.8))
		isOK, v := sas.doCheckRule(rule)
		assert.Equal(t, true, isOK)
		assert.Equal(t, float64(0.8), v)
		currentCpuUsage.Store(float64(notRetrievedValue))
	})
}

func TestDoCheckRuleDefault(t *testing.T) {
	var sas *SystemAdaptiveSlot
	rule := &SystemRule{MetricType: MetricTypeSize,
		TriggerCount: 0.5}

	isOK, v := sas.doCheckRule(rule)
	assert.Equal(t, true, isOK)
	assert.Equal(t, float64(0), v)
}

func TestString(t *testing.T) {
	var sas *SystemAdaptiveSlot

	assert.True(t, sas.String() == SlotName)
}
