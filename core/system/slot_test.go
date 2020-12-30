// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func TestCheckNilInput(t *testing.T) {
	var sas *AdaptiveSlot

	t.Run("NilInput", func(t *testing.T) {
		r := sas.Check(nil)
		assert.True(t, r == nil)
	})

	t.Run("NilResourceInput", func(t *testing.T) {
		r := sas.Check(&base.EntryContext{})
		assert.True(t, r == nil)
	})

	t.Run("UnsuitableFlowType", func(t *testing.T) {
		rw := base.NewResourceWrapper("test", base.ResTypeCommon, base.Outbound)
		r := sas.Check(&base.EntryContext{Resource: rw})
		assert.True(t, r == nil)
	})
}

func TestCheckEmptyRule(t *testing.T) {
	var sas *AdaptiveSlot
	rw := base.NewResourceWrapper("test", base.ResTypeCommon, base.Inbound)
	r := sas.Check(&base.EntryContext{
		Resource:        rw,
		RuleCheckResult: base.NewTokenResultPass(),
	})
	assert.True(t, r == nil || r.IsPass())
}

func TestDoCheckRuleConcurrency(t *testing.T) {
	var sas *AdaptiveSlot
	rule := &Rule{MetricType: Concurrency,
		TriggerCount: 0.5}

	t.Run("TrueConcurrency", func(t *testing.T) {
		isOK, _, v := sas.doCheckRule(rule)
		assert.True(t, util.Float64Equals(float64(0.0), v))
		assert.Equal(t, true, isOK)
	})

	t.Run("FalseConcurrency", func(t *testing.T) {
		stat.InboundNode().IncreaseConcurrency()
		isOK, _, v := sas.doCheckRule(rule)
		assert.True(t, util.Float64Equals(float64(1.0), v))
		assert.Equal(t, false, isOK)
		stat.InboundNode().DecreaseConcurrency()
	})
}

func TestDoCheckRuleLoad(t *testing.T) {
	var sas *AdaptiveSlot
	rule := &Rule{MetricType: Load,
		TriggerCount: 0.5}

	t.Run("TrueLoad", func(t *testing.T) {
		isOK, _, v := sas.doCheckRule(rule)
		assert.True(t, util.Float64Equals(system_metric.NotRetrievedLoadValue, v))
		assert.Equal(t, true, isOK)
	})

	t.Run("BBRTrueLoad", func(t *testing.T) {
		rule.Strategy = BBR
		system_metric.SetSystemLoad(1.0)
		isOK, _, v := sas.doCheckRule(rule)
		assert.Equal(t, true, isOK)
		assert.True(t, util.Float64Equals(float64(1.0), v))
		system_metric.SetSystemLoad(system_metric.NotRetrievedLoadValue)
	})
}

func TestDoCheckRuleCpuUsage(t *testing.T) {
	var sas *AdaptiveSlot
	rule := &Rule{
		MetricType:   CpuUsage,
		TriggerCount: 0.5,
	}

	t.Run("TrueCpuUsage", func(t *testing.T) {
		isOK, _, v := sas.doCheckRule(rule)
		assert.True(t, util.Float64Equals(system_metric.NotRetrievedCpuUsageValue, v))
		assert.Equal(t, true, isOK)
	})

	t.Run("BBRTrueCpuUsage", func(t *testing.T) {
		rule.Strategy = BBR
		system_metric.SetSystemCpuUsage(0.8)
		isOK, _, v := sas.doCheckRule(rule)
		assert.Equal(t, true, isOK)
		assert.True(t, util.Float64Equals(0.8, v))
		system_metric.SetSystemCpuUsage(system_metric.NotRetrievedCpuUsageValue)
	})
}

func TestDoCheckRuleDefault(t *testing.T) {
	var sas *AdaptiveSlot
	rule := &Rule{MetricType: MetricTypeSize,
		TriggerCount: 0.5}

	isOK, _, v := sas.doCheckRule(rule)
	assert.Equal(t, true, isOK)
	assert.True(t, util.Float64Equals(float64(0.0), v))
}
