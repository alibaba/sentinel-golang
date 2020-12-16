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

	"github.com/stretchr/testify/assert"
)

func TestGetRules(t *testing.T) {
	t.Run("EmptyRules", func(t *testing.T) {
		rules := getRules()
		assert.Equal(t, 0, len(rules))
	})

	t.Run("GetUpdatedRules", func(t *testing.T) {
		defer func() { ruleMap = make(RuleMap) }()

		r := map[MetricType][]*Rule{
			InboundQPS:  {&Rule{MetricType: InboundQPS, TriggerCount: 1}},
			Concurrency: {&Rule{MetricType: Concurrency, TriggerCount: 2}},
		}
		ruleMap = r
		rules := getRules()
		assert.Equal(t, 2, len(rules))

		r[InboundQPS] = append(r[InboundQPS], &Rule{MetricType: InboundQPS, TriggerCount: 2})
		ruleMap = r
		rules = getRules()
		assert.Equal(t, 3, len(rules))
	})
}

func TestLoadRules(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		isOK, err := LoadRules(nil)
		assert.Equal(t, true, isOK)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(ruleMap))
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		defer func() { ruleMap = make(RuleMap) }()
		sRule := []*Rule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		isOK, err := LoadRules(sRule)
		assert.Equal(t, true, isOK)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(ruleMap))
	})
}

func TestClearRules(t *testing.T) {
	t.Run("EmptyOriginRuleMap", func(t *testing.T) {
		err := ClearRules()
		assert.Equal(t, 0, len(ruleMap))
		assert.Nil(t, err)
	})

	t.Run("NoEmptyOriginRuleMap", func(t *testing.T) {
		r := []*Rule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		isOK, err := LoadRules(r)
		assert.Equal(t, true, isOK)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(ruleMap))
		err = ClearRules()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(ruleMap))
	})
}

func TestOnRuleUpdate(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		err := onRuleUpdate(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ruleMap))
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		defer func() { ruleMap = make(RuleMap) }()
		rMap := RuleMap{
			InboundQPS: []*Rule{
				{MetricType: InboundQPS, TriggerCount: 1},
			},
			Concurrency: []*Rule{
				{MetricType: Concurrency, TriggerCount: 2},
			},
		}
		err := onRuleUpdate(rMap)
		assert.NoError(t, err)
		assert.Equal(t, len(rMap), len(ruleMap))
	})
}

func TestBuildRuleMap(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		r := buildRuleMap(nil)
		assert.Equal(t, 0, len(r))
	})

	t.Run("InvalidSystemRule", func(t *testing.T) {
		sRule := []*Rule{
			{MetricType: InboundQPS, TriggerCount: -1},
		}
		r := buildRuleMap(sRule)
		assert.Equal(t, 0, len(r))
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		sRule := []*Rule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		r := buildRuleMap(sRule)
		assert.Equal(t, len(sRule), len(r))
	})

	t.Run("MultiRuleOneTypeValidSystemRule", func(t *testing.T) {
		sRule := []*Rule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: InboundQPS, TriggerCount: 2},
		}
		r := buildRuleMap(sRule)
		assert.Equal(t, 1, len(r))
	})
}

func TestIsValidSystemRule(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		err := IsValidSystemRule(nil)
		assert.EqualError(t, err, "nil Rule")
	})

	t.Run("NegativeThreshold", func(t *testing.T) {
		sRule := &Rule{MetricType: InboundQPS, TriggerCount: -1}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "negative threshold")
	})

	t.Run("InvalidMetricType", func(t *testing.T) {
		sRule := &Rule{MetricType: MetricTypeSize}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "invalid metric type")
	})

	t.Run("InvalidCPUUsage", func(t *testing.T) {
		sRule := &Rule{MetricType: CpuUsage, TriggerCount: 75}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "invalid CPU usage, valid range is [0.0, 1.0]")
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		sRule := &Rule{MetricType: Load, TriggerCount: 12, Strategy: BBR}
		err := IsValidSystemRule(sRule)
		assert.NoError(t, err)
	})
}
