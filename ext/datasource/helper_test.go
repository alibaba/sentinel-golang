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
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFlowRulesJsonConverter(t *testing.T) {
	// Prepare test data
	f, err := os.Open("../../tests/testdata/extension/helper/FlowRule.json")
	defer f.Close()
	if err != nil {
		t.Errorf("The rules file is not existed, err:%+v.", errors.WithStack(err))
	}
	normalSrc, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
	}

	t.Run("TestFlowRulesJsonConverter_nil", func(t *testing.T) {
		got, err := FlowRuleJsonArrayParser(nil)
		assert.True(t, got == nil && err == nil)

		got, err = FlowRuleJsonArrayParser([]byte{})
		assert.True(t, got == nil && err == nil)
	})

	t.Run("TestFlowRulesJsonConverter_error", func(t *testing.T) {
		_, err := FlowRuleJsonArrayParser([]byte{'x', 'i', 'm', 'u'})
		assert.True(t, err != nil)
	})

	t.Run("TestFlowRulesJsonConverter_normal", func(t *testing.T) {
		got, err := FlowRuleJsonArrayParser(normalSrc)
		assert.True(t, got != nil && err == nil)
		flowRules := got.([]*flow.Rule)
		assert.True(t, len(flowRules) == 3)
		r1 := &flow.Rule{
			Resource:               "abc",
			MetricType:             flow.Concurrency,
			Count:                  100,
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
			MetricType:             flow.QPS,
			Count:                  200,
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
			MetricType:             flow.QPS,
			Count:                  300,
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
				ID:                     0,
				Resource:               "abc",
				MetricType:             0,
				Count:                  0,
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
			ID:                     0,
			Resource:               "aaaa",
			MetricType:             0,
			Count:                  0,
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

func TestSystemRulesJsonConvert(t *testing.T) {
	// Prepare test data
	f, err := os.Open("../../tests/testdata/extension/helper/SystemRule.json")
	defer f.Close()
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
}

func TestSystemRulesUpdater(t *testing.T) {
	t.Run("TestSystemRulesUpdater_Nil", func(t *testing.T) {
		system.ClearRules()
		system.LoadRules([]*system.Rule{
			{
				ID:           0,
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
			ID:           0,
			MetricType:   0,
			TriggerCount: 0,
			Strategy:     0,
		}
		p = append(p, sr)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 1)
	})
}

func TestCircuitBreakerRulesJsonConverter(t *testing.T) {
	t.Run("TestCircuitBreakerRulesJsonConverter_failed", func(t *testing.T) {
		_, err := CircuitBreakerRuleJsonArrayParser([]byte{'s', 'r', 'c'})
		assert.True(t, err != nil)
	})

	t.Run("TestCircuitBreakerRulesJsonConverter_Succeed", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/CircuitBreakerRule.json")
		defer f.Close()
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
}

func TestCircuitBreakerRulesUpdater(t *testing.T) {
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

	rules := cb.GetResRules("abc")
	assert.True(t, reflect.DeepEqual(rules[0], r1))
	assert.True(t, reflect.DeepEqual(rules[1], r2))
	assert.True(t, reflect.DeepEqual(rules[2], r3))
}

func TestHotSpotParamRuleListJsonConverter(t *testing.T) {
	t.Run("TestHotSpotParamRuleListJsonConverter_invalid", func(t *testing.T) {
		properties, err := HotSpotParamRuleJsonArrayParser([]byte{'s', 'r', 'c'})
		assert.True(t, properties == nil)
		assert.True(t, err != nil)
	})

	t.Run("TestHotSpotParamRuleListJsonConverter_normal", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/HotSpotParamFlowRule.json")
		defer f.Close()
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
		assert.True(t, strings.Contains(rules[0].String(), "Resource:abc, MetricType:Concurrency, ControlBehavior:Reject, ParamIndex:0, Threshold:1000.000000, MaxQueueingTimeMs:1, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:[{ValKind:KindInt ValStr:1000 Threshold:10001} {ValKind:KindString ValStr:ximu Threshold:10002} {ValKind:KindBool ValStr:true Threshold:10003}]}"))
		assert.True(t, strings.Contains(rules[1].String(), "Resource:abc, MetricType:Concurrency, ControlBehavior:Throttling, ParamIndex:1, Threshold:2000.000000, MaxQueueingTimeMs:2, BurstCount:20, DurationInSec:2, ParamsMaxCapacity:20000, SpecificItems:[{ValKind:KindInt ValStr:1000 Threshold:20001} {ValKind:KindString ValStr:ximu Threshold:20002} {ValKind:KindBool ValStr:true Threshold:20003}]}"))
		assert.True(t, strings.Contains(rules[2].String(), "Resource:abc, MetricType:QPS, ControlBehavior:Reject, ParamIndex:2, Threshold:3000.000000, MaxQueueingTimeMs:3, BurstCount:30, DurationInSec:3, ParamsMaxCapacity:30000, SpecificItems:[{ValKind:KindInt ValStr:1000 Threshold:30001} {ValKind:KindString ValStr:ximu Threshold:30002} {ValKind:KindBool ValStr:true Threshold:30003}]}"))
		assert.True(t, strings.Contains(rules[3].String(), "Resource:abc, MetricType:QPS, ControlBehavior:Throttling, ParamIndex:3, Threshold:4000.000000, MaxQueueingTimeMs:4, BurstCount:40, DurationInSec:4, ParamsMaxCapacity:40000, SpecificItems:[{ValKind:KindInt ValStr:1000 Threshold:40001} {ValKind:KindString ValStr:ximu Threshold:40002} {ValKind:KindBool ValStr:true Threshold:40003}]}"))
	})
}

func TestHotSpotParamRuleListJsonUpdater(t *testing.T) {
	// Prepare test data
	m := make([]hotspot.SpecificValue, 2)
	m[0] = hotspot.SpecificValue{
		ValKind:   hotspot.KindString,
		ValStr:    "sss",
		Threshold: 1,
	}
	m[1] = hotspot.SpecificValue{
		ValKind:   hotspot.KindFloat64,
		ValStr:    "1.123",
		Threshold: 3,
	}
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

	m2 := make([]hotspot.SpecificValue, 2)
	m2[0] = hotspot.SpecificValue{
		ValKind:   hotspot.KindString,
		ValStr:    "sss",
		Threshold: 1,
	}
	m2[1] = hotspot.SpecificValue{
		ValKind:   hotspot.KindFloat64,
		ValStr:    "1.123",
		Threshold: 3,
	}
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

	m3 := make([]hotspot.SpecificValue, 2)
	m3[0] = hotspot.SpecificValue{
		ValKind:   hotspot.KindString,
		ValStr:    "sss",
		Threshold: 1,
	}
	m3[1] = hotspot.SpecificValue{
		ValKind:   hotspot.KindFloat64,
		ValStr:    "1.123",
		Threshold: 3,
	}
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

	rules := hotspot.GetRules("abc")
	assert.True(t, rules[0].Equals(r1))
	assert.True(t, rules[1].Equals(r2))
	assert.True(t, rules[2].Equals(r3))
	assert.True(t, rules[3].Equals(r4))
}
