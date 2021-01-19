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

package datasource

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	cb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFlowRuleJsonArrayParser(t *testing.T) {
	// Prepare test data
	f, err := os.Open("../../tests/testdata/extension/helper/FlowRule.json")
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Errorf("The rules file is not existed, err:%+v.", errors.WithStack(err))
	}
	normalSrc, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
	}

	t.Run("TestFlowRuleJsonArrayParser_Nil", func(t *testing.T) {
		got, err := FlowRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = FlowRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})

	t.Run("TestFlowRuleJsonArrayParser_Error", func(t *testing.T) {
		_, err := FlowRuleJsonArrayParser([]byte{'x', 'i', 'm', 'u'})
		assert.True(t, err != nil)
	})

	t.Run("TestFlowRuleJsonArrayParser_Normal", func(t *testing.T) {
		got, err := FlowRuleJsonArrayParser(normalSrc)
		assert.True(t, got != nil && err == nil)
		flowRules := got.([]*flow.Rule)
		assert.True(t, len(flowRules) == 3)
		r1 := &flow.Rule{
			Resource:               "abc",
			Threshold:              100,
			RelationStrategy:       flow.CurrentResource,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			RefResource:            "refDefault",
			WarmUpPeriodSec:        10,
			MaxQueueingTimeMs:      1000,
		}
		assert.True(t, reflect.DeepEqual(flowRules[0], r1))

		r2 := &flow.Rule{
			Resource:               "abc",
			Threshold:              200,
			RelationStrategy:       flow.AssociatedResource,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Throttling,
			RefResource:            "refDefault",
			WarmUpPeriodSec:        20,
			MaxQueueingTimeMs:      2000,
		}
		assert.True(t, reflect.DeepEqual(flowRules[1], r2))

		r3 := &flow.Rule{
			Resource:               "abc",
			Threshold:              300,
			RelationStrategy:       flow.CurrentResource,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Throttling,
			RefResource:            "refDefault",
			WarmUpPeriodSec:        30,
			MaxQueueingTimeMs:      3000,
		}
		assert.True(t, reflect.DeepEqual(flowRules[2], r3))
	})
}

func TestFlowRulesUpdater(t *testing.T) {
	t.Run("TestFlowRulesUpdater_Nil", func(t *testing.T) {
		flow.ClearRules()
		flow.LoadRules([]*flow.Rule{
			{
				Resource:               "abc",
				Threshold:              0,
				RelationStrategy:       0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				RefResource:            "",
				WarmUpPeriodSec:        0,
				MaxQueueingTimeMs:      0,
			}})
		assert.True(t, len(flow.GetRules()) == 1, "Fail to prepare test data.")
		err := FlowRulesUpdater(nil)
		assert.True(t, err == nil && len(flow.GetRules()) == 0, "Fail to test TestFlowRulesUpdater_Nil")
	})

	t.Run("TestFlowRulesUpdater_Assert_Failed", func(t *testing.T) {
		flow.ClearRules()
		err := FlowRulesUpdater("xxxxxxxx")
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []flow.Rule"))
	})

	t.Run("TestFlowRulesUpdater_Empty_Rules", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.Rule, 0)
		err := FlowRulesUpdater(p)
		assert.True(t, err == nil && len(flow.GetRules()) == 0)
	})

	t.Run("TestFlowRulesUpdater_Normal", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.Rule, 0)
		fw := flow.Rule{
			Resource:               "aaaa",
			Threshold:              0,
			RelationStrategy:       0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		p = append(p, fw)
		err := FlowRulesUpdater(p)
		assert.True(t, err == nil && len(flow.GetRules()) == 1)
	})
}

func TestSystemRuleJsonArrayParser(t *testing.T) {
	t.Run("TestSystemRuleJsonArrayParser_Normal", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/SystemRule.json")
		defer func() {
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}()
		if err != nil {
			t.Errorf("The rules file is not existed, err:%+v.", errors.WithStack(err))
		}
		normalSrc, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
		}

		got, err := SystemRuleJsonArrayParser(normalSrc)
		systemRules := got.([]*system.Rule)
		assert.True(t, err == nil && len(systemRules) == 4)

		r0 := &system.Rule{
			MetricType:   system.Load,
			TriggerCount: 0.5,
			Strategy:     system.BBR,
		}
		r1 := &system.Rule{
			MetricType:   system.AvgRT,
			TriggerCount: 0.6,
			Strategy:     system.BBR,
		}
		r2 := &system.Rule{
			MetricType:   system.Concurrency,
			TriggerCount: 0.7,
			Strategy:     system.BBR,
		}
		r3 := &system.Rule{
			MetricType:   system.InboundQPS,
			TriggerCount: 0.8,
			Strategy:     system.BBR,
		}

		assert.True(t, reflect.DeepEqual(r0, systemRules[0]))
		assert.True(t, reflect.DeepEqual(r1, systemRules[1]))
		assert.True(t, reflect.DeepEqual(r2, systemRules[2]))
		assert.True(t, reflect.DeepEqual(r3, systemRules[3]))
	})

	t.Run("TestSystemRuleJsonArrayParser_Nil", func(t *testing.T) {
		got, err := SystemRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = SystemRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})
}

func TestSystemRulesUpdater(t *testing.T) {
	t.Run("TestSystemRulesUpdater_Nil", func(t *testing.T) {
		system.ClearRules()
		system.LoadRules([]*system.Rule{
			{
				MetricType:   0,
				TriggerCount: 0,
				Strategy:     0,
			},
		})
		assert.True(t, len(system.GetRules()) == 1, "Fail to prepare data.")
		err := SystemRulesUpdater(nil)
		assert.True(t, err == nil && len(system.GetRules()) == 0, "Fail to test TestSystemRulesUpdater_Nil")
	})

	t.Run("TestSystemRulesUpdater_Assert_Failed", func(t *testing.T) {
		system.ClearRules()
		err := SystemRulesUpdater("xxxxxxxx")
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []system.Rule"))
	})

	t.Run("TestSystemRulesUpdater_Empty_Rules", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.Rule, 0)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 0)
	})

	t.Run("TestSystemRulesUpdater_Normal", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.Rule, 0)
		sr := system.Rule{
			MetricType:   0,
			TriggerCount: 0,
			Strategy:     0,
		}
		p = append(p, sr)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 1)
	})
}

func TestCircuitBreakerRuleJsonArrayParser(t *testing.T) {
	t.Run("TestCircuitBreakerRuleJsonArrayParser_Failed", func(t *testing.T) {
		_, err := CircuitBreakerRuleJsonArrayParser([]byte{'s', 'r', 'c'})
		assert.True(t, err != nil)
	})

	t.Run("TestCircuitBreakerRuleJsonArrayParser_Succeed", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/CircuitBreakerRule.json")
		defer func() {
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}()
		if err != nil {
			t.Errorf("The rules file is not existed, err:%+v.", err)
		}
		src, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("Fail to read file, err: %+v.", err)
		}

		properties, err := CircuitBreakerRuleJsonArrayParser(src)
		rules := properties.([]*cb.Rule)
		assert.True(t, err == nil)
		assert.True(t, len(rules) == 3)
		assert.True(t, reflect.DeepEqual(rules[0], &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   10,
			MinRequestAmount: 10,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   100,
			Threshold:        0.1,
		}))
		assert.True(t, reflect.DeepEqual(rules[1], &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.ErrorRatio,
			RetryTimeoutMs:   20,
			MinRequestAmount: 20,
			StatIntervalMs:   2000,
			Threshold:        0.2,
		}))
		assert.True(t, reflect.DeepEqual(rules[2], &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.ErrorCount,
			RetryTimeoutMs:   30,
			MinRequestAmount: 30,
			StatIntervalMs:   3000,
			Threshold:        30,
		}))
	})

	t.Run("TestCircuitBreakerRuleJsonArrayParser_Nil", func(t *testing.T) {
		got, err := CircuitBreakerRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = CircuitBreakerRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})
}

func TestCircuitBreakerRulesUpdater(t *testing.T) {
	t.Run("TestCircuitBreakerRulesUpdater_Normal", func(t *testing.T) {
		// Prepare test data
		r1 := &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		}
		r2 := &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.ErrorRatio,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        0.3,
		}
		r3 := &cb.Rule{
			Resource:         "abc",
			Strategy:         cb.ErrorCount,
			RetryTimeoutMs:   1000,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			Threshold:        10,
		}

		err := CircuitBreakerRulesUpdater([]*cb.Rule{r1, r2, r3})
		assert.True(t, err == nil)

		rules := cb.GetRulesOfResource("abc")
		assert.True(t, reflect.DeepEqual(rules[0], *r1))
		assert.True(t, reflect.DeepEqual(rules[1], *r2))
		assert.True(t, reflect.DeepEqual(rules[2], *r3))
	})
	t.Run("TestCircuitBreakerRulesUpdater_Nil", func(t *testing.T) {
		err := CircuitBreakerRulesUpdater(nil)
		assert.Nil(t, err)
	})
	t.Run("TestCircuitBreakerRulesUpdater_Type_Err", func(t *testing.T) {
		rules := []*flow.Rule{
			{
				Resource:               "abc",
				Threshold:              0,
				RelationStrategy:       0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				RefResource:            "",
				WarmUpPeriodSec:        0,
				MaxQueueingTimeMs:      0,
			}}
		err := CircuitBreakerRulesUpdater(rules)

		assert.True(t, err.(Error).Code() == UpdatePropertyError)
		assert.True(t, strings.Contains(err.(Error).desc, "Fail to type assert"))
	})
}

func TestHotSpotParamRuleJsonArrayParser(t *testing.T) {
	t.Run("TestHotSpotParamJsonArrayParser_Invalid", func(t *testing.T) {
		_, err := HotSpotParamRuleJsonArrayParser([]byte{'s', 'r', 'c'})
		assert.True(t, err != nil)
	})

	t.Run("TestHotSpotParamRuleJsonArrayParser_Normal", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/HotSpotParamFlowRule.json")
		defer func() {
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}()
		if err != nil {
			t.Errorf("The rules file is not existed, err:%+v.", err)
		}
		src, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("Fail to read file, err: %+v.", err)
		}

		properties, err := HotSpotParamRuleJsonArrayParser(src)
		rules := properties.([]*hotspot.Rule)
		assert.True(t, err == nil)
		assert.True(t, len(rules) == 4)
		for _, r := range rules {
			fmt.Println(r)
		}
		assert.True(t, strings.Contains(rules[0].String(), "Resource:abc, MetricType:Concurrency, ControlBehavior:Reject, ParamIndex:0, ParamKey:, Threshold:1000, MaxQueueingTimeMs:1, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:map[true:10003 1000:10001 ximu:10002]"))
		assert.True(t, strings.Contains(rules[1].String(), "Resource:abc, MetricType:Concurrency, ControlBehavior:Throttling, ParamIndex:1, ParamKey:, Threshold:2000, MaxQueueingTimeMs:2, BurstCount:20, DurationInSec:2, ParamsMaxCapacity:20000, SpecificItems:map[true:20003 1000:20001 ximu:20002"))
		assert.True(t, strings.Contains(rules[2].String(), "Resource:abc, MetricType:QPS, ControlBehavior:Reject, ParamIndex:2, ParamKey:, Threshold:3000, MaxQueueingTimeMs:3, BurstCount:30, DurationInSec:3, ParamsMaxCapacity:30000, SpecificItems:map[true:30003 1000:30001 ximu:30002"))
		assert.True(t, strings.Contains(rules[3].String(), "Resource:abc, MetricType:QPS, ControlBehavior:Throttling, ParamIndex:3, ParamKey:, Threshold:4000, MaxQueueingTimeMs:4, BurstCount:40, DurationInSec:4, ParamsMaxCapacity:40000, SpecificItems:map[true:40003 1000:40001 ximu:40002"))
	})

	t.Run("TestHotSpotParamRuleJsonArrayParser_Nil", func(t *testing.T) {
		got, err := HotSpotParamRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = HotSpotParamRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})
}

func TestHotSpotParamRuleListJsonUpdater(t *testing.T) {
	t.Run("TestHotSpotParamRuleListJsonUpdater", func(t *testing.T) {
		// Prepare test data
		m := make(map[interface{}]int64)
		r1 := &hotspot.Rule{
			ID:                "1",
			Resource:          "abc",
			MetricType:        hotspot.Concurrency,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0,
			Threshold:         100,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     m,
		}

		m2 := make(map[interface{}]int64)
		r2 := &hotspot.Rule{
			ID:                "2",
			Resource:          "abc",
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        1,
			Threshold:         100,
			MaxQueueingTimeMs: 20,
			BurstCount:        0,
			DurationInSec:     1,
			SpecificItems:     m2,
		}

		m3 := make(map[interface{}]int64)
		r3 := &hotspot.Rule{
			ID:                "3",
			Resource:          "abc",
			MetricType:        hotspot.Concurrency,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        2,
			Threshold:         100,
			MaxQueueingTimeMs: 20,
			BurstCount:        0,
			DurationInSec:     1,
			SpecificItems:     m3,
		}

		r4 := &hotspot.Rule{
			ID:                "4",
			Resource:          "abc",
			MetricType:        hotspot.Concurrency,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        2,
			Threshold:         100,
			MaxQueueingTimeMs: 20,
			BurstCount:        0,
			DurationInSec:     2,
			SpecificItems:     m3,
		}

		err := HotSpotParamRulesUpdater([]*hotspot.Rule{r1, r2, r3, r4})
		assert.True(t, err == nil)

		rules := hotspot.GetRulesOfResource("abc")
		assert.True(t, reflect.DeepEqual(rules[0], *r1))
		assert.True(t, reflect.DeepEqual(rules[1], *r2))
		assert.True(t, reflect.DeepEqual(rules[2], *r3))
		assert.True(t, reflect.DeepEqual(rules[3], *r4))
	})

	t.Run("TestHotSpotParamRuleListJsonUpdater_Type_Err", func(t *testing.T) {
		rules := []*flow.Rule{
			{
				Resource:               "abc",
				Threshold:              0,
				RelationStrategy:       0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				RefResource:            "",
				WarmUpPeriodSec:        0,
				MaxQueueingTimeMs:      0,
			}}
		err := HotSpotParamRulesUpdater(rules)

		assert.True(t, err.(Error).Code() == UpdatePropertyError)
		assert.True(t, strings.Contains(err.(Error).desc, "Fail to type assert"))
	})
}

func TestIsolationRuleJsonArrayParser(t *testing.T) {
	t.Run("TestIsolationJsonArrayParser_Invalid", func(t *testing.T) {
		_, err := IsolationRuleJsonArrayParser([]byte{'s', 'r', 'c'})
		assert.True(t, err != nil)
	})

	t.Run("TestIsolationRuleJsonArrayParser_Normal", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/IsolationRule.json")
		defer func() {
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}()
		if err != nil {
			t.Errorf("The rules file is not existed, err:%+v.", err)
		}
		src, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("Fail to read file, err: %+v.", err)
		}

		properties, err := IsolationRuleJsonArrayParser(src)
		rules := properties.([]*isolation.Rule)
		assert.True(t, err == nil)
		assert.True(t, len(rules) == 4)
		assert.True(t, strings.Contains(rules[0].String(), `{"resource":"abc","metricType":0,"threshold":100}`))
		assert.True(t, strings.Contains(rules[1].String(), `{"resource":"abc","metricType":0,"threshold":90}`))
		assert.True(t, strings.Contains(rules[2].String(), `{"resource":"abc","metricType":0,"threshold":80}`))
		assert.True(t, strings.Contains(rules[3].String(), `{"resource":"abc","metricType":0,"threshold":70}`))
	})

	t.Run("TestIsolationRuleJsonArrayParser_Nil", func(t *testing.T) {
		got, err := IsolationRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = IsolationRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})
}

func TestIsolationRuleListJsonUpdater(t *testing.T) {
	t.Run("TestIsolationRuleListJsonUpdater", func(t *testing.T) {
		// Prepare test data
		r1 := &isolation.Rule{
			Resource:   "abc",
			MetricType: isolation.Concurrency,
			Threshold:  100,
		}

		r2 := &isolation.Rule{
			Resource:   "abc",
			MetricType: isolation.Concurrency,
			Threshold:  90,
		}

		r3 := &isolation.Rule{
			Resource:   "abc",
			MetricType: isolation.Concurrency,
			Threshold:  80,
		}

		r4 := &isolation.Rule{
			Resource:   "abc",
			MetricType: isolation.Concurrency,
			Threshold:  70,
		}

		err := IsolationRulesUpdater([]*isolation.Rule{r1, r2, r3, r4})
		assert.True(t, err == nil)

		rules := isolation.GetRulesOfResource("abc")
		assert.True(t, reflect.DeepEqual(rules[0], *r1))
		assert.True(t, reflect.DeepEqual(rules[1], *r2))
		assert.True(t, reflect.DeepEqual(rules[2], *r3))
		assert.True(t, reflect.DeepEqual(rules[3], *r4))
	})

	t.Run("TestIsolationRuleListJsonUpdater_Type_Err", func(t *testing.T) {
		rules := []*flow.Rule{
			{
				Resource:               "abc",
				Threshold:              0,
				RelationStrategy:       0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				RefResource:            "",
				WarmUpPeriodSec:        0,
				MaxQueueingTimeMs:      0,
			}}
		err := IsolationRulesUpdater(rules)

		assert.True(t, err.(Error).Code() == UpdatePropertyError)
		assert.True(t, strings.Contains(err.(Error).desc, "Fail to type assert"))
	})
}
