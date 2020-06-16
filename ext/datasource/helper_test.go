package datasource

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/freq_params_traffic"
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
		got, err := FlowRulesJsonConverter(nil)
		assert.True(t, got == nil && err == nil)

		got, err = FlowRulesJsonConverter([]byte{})
		assert.True(t, got == nil && err == nil)
	})

	t.Run("TestFlowRulesJsonConverter_error", func(t *testing.T) {
		got, err := FlowRulesJsonConverter([]byte{'x', 'i', 'm', 'u'})
		assert.True(t, got == nil)
		realErr, succ := err.(Error)
		assert.True(t, succ && realErr.code == ConvertSourceError)
	})

	t.Run("TestFlowRulesJsonConverter_normal", func(t *testing.T) {
		got, _ := FlowRulesJsonConverter(normalSrc)
		assert.True(t, got != nil)
		flowRules := got.([]*flow.FlowRule)
		assert.True(t, len(flowRules) == 3)
		r1 := &flow.FlowRule{
			Resource:          "abc",
			LimitOrigin:       "default",
			MetricType:        flow.Concurrency,
			Count:             100,
			RelationStrategy:  flow.Direct,
			ControlBehavior:   flow.Reject,
			RefResource:       "refDefault",
			WarmUpPeriodSec:   10,
			MaxQueueingTimeMs: 1000,
			ClusterMode:       false,
			ClusterConfig: flow.ClusterRuleConfig{
				ThresholdType: flow.AvgLocalThreshold,
			},
		}
		assert.True(t, reflect.DeepEqual(flowRules[0], r1))

		r2 := &flow.FlowRule{
			Resource:          "abc",
			LimitOrigin:       "default",
			MetricType:        flow.QPS,
			Count:             200,
			RelationStrategy:  flow.AssociatedResource,
			ControlBehavior:   flow.WarmUp,
			RefResource:       "refDefault",
			WarmUpPeriodSec:   20,
			MaxQueueingTimeMs: 2000,
			ClusterMode:       true,
			ClusterConfig: flow.ClusterRuleConfig{
				ThresholdType: flow.GlobalThreshold,
			},
		}
		assert.True(t, reflect.DeepEqual(flowRules[1], r2))

		r3 := &flow.FlowRule{
			Resource:          "abc",
			LimitOrigin:       "default",
			MetricType:        flow.QPS,
			Count:             300,
			RelationStrategy:  flow.Direct,
			ControlBehavior:   flow.WarmUp,
			RefResource:       "refDefault",
			WarmUpPeriodSec:   30,
			MaxQueueingTimeMs: 3000,
			ClusterMode:       true,
			ClusterConfig: flow.ClusterRuleConfig{
				ThresholdType: flow.GlobalThreshold,
			},
		}
		assert.True(t, reflect.DeepEqual(flowRules[2], r3))
	})
}

func TestFlowRulesUpdater(t *testing.T) {
	t.Run("TestFlowRulesUpdater_Nil", func(t *testing.T) {
		flow.ClearRules()
		flow.LoadRules([]*flow.FlowRule{
			{
				ID:                0,
				Resource:          "abc",
				LimitOrigin:       "default",
				MetricType:        0,
				Count:             0,
				RelationStrategy:  0,
				ControlBehavior:   0,
				RefResource:       "",
				WarmUpPeriodSec:   0,
				MaxQueueingTimeMs: 0,
				ClusterMode:       false,
				ClusterConfig:     flow.ClusterRuleConfig{},
			}})
		assert.True(t, len(flow.GetRules()) == 1, "Fail to prepare test data.")
		err := FlowRulesUpdater(nil)
		assert.True(t, err == nil && len(flow.GetRules()) == 0, "Fail to test TestFlowRulesUpdater_Nil")
	})

	t.Run("TestFlowRulesUpdater_Assert_Failed", func(t *testing.T) {
		flow.ClearRules()
		err := FlowRulesUpdater("xxxxxxxx")
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []flow.FlowRule"))
	})

	t.Run("TestFlowRulesUpdater_Empty_Rules", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.FlowRule, 0)
		err := FlowRulesUpdater(p)
		assert.True(t, err == nil && len(flow.GetRules()) == 0)
	})

	t.Run("TestFlowRulesUpdater_Normal", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.FlowRule, 0)
		fw := flow.FlowRule{
			ID:                0,
			Resource:          "aaaa",
			LimitOrigin:       "aaa",
			MetricType:        0,
			Count:             0,
			RelationStrategy:  0,
			ControlBehavior:   0,
			RefResource:       "",
			WarmUpPeriodSec:   0,
			MaxQueueingTimeMs: 0,
			ClusterMode:       false,
			ClusterConfig:     flow.ClusterRuleConfig{},
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

	got, err := SystemRulesJsonConverter(normalSrc)
	systemRules := got.([]*system.SystemRule)
	assert.True(t, err == nil && len(systemRules) == 4)

	r0 := &system.SystemRule{
		MetricType:   system.Load,
		TriggerCount: 0.5,
		Strategy:     system.BBR,
	}
	r1 := &system.SystemRule{
		MetricType:   system.AvgRT,
		TriggerCount: 0.6,
		Strategy:     system.BBR,
	}
	r2 := &system.SystemRule{
		MetricType:   system.Concurrency,
		TriggerCount: 0.7,
		Strategy:     system.BBR,
	}
	r3 := &system.SystemRule{
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
		system.LoadRules([]*system.SystemRule{
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
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []system.SystemRule"))
	})

	t.Run("TestSystemRulesUpdater_Empty_Rules", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.SystemRule, 0)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 0)
	})

	t.Run("TestSystemRulesUpdater_Normal", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.SystemRule, 0)
		sr := system.SystemRule{
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
		properties, err := CircuitBreakerRulesJsonConverter([]byte{'s', 'r', 'c'})
		assert.True(t, properties == nil)
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

		properties, err := CircuitBreakerRulesJsonConverter(src)
		rules := properties.([]circuitbreaker.Rule)
		assert.True(t, err == nil)
		assert.True(t, len(rules) == 3)
		assert.True(t, strings.Contains(rules[0].String(), "resource=abc, strategy=SlowRequestRatio, RetryTimeoutMs=10, MinRequestAmount=10}, MaxAllowedRt=100, MaxSlowRequestRatio=0.100000"))
		assert.True(t, strings.Contains(rules[1].String(), "resource=abc, strategy=ErrorRatio, RetryTimeoutMs=20, MinRequestAmount=20}, Threshold=0.200000"))
		assert.True(t, strings.Contains(rules[2].String(), "resource=abc, strategy=ErrorCount, RetryTimeoutMs=30, MinRequestAmount=30}, Threshold=30"))
	})
}

func TestCircuitBreakerRulesUpdater(t *testing.T) {
	// Prepare test data
	r1 := circuitbreaker.NewSlowRtRule("abc", 1000, 1, 20, 5, 0.1)
	r2 := circuitbreaker.NewErrorRatioRule("abc", 1000, 1, 5, 0.3)
	r3 := circuitbreaker.NewErrorCountRule("abc", 1000, 1, 5, 10)

	err := CircuitBreakerRulesUpdater([]circuitbreaker.Rule{r1, r2, r3})
	assert.True(t, err == nil)

	rules := circuitbreaker.GetResRules("abc")
	assert.True(t, rules[0].IsEqualsTo(r1))
	assert.True(t, rules[1].IsEqualsTo(r2))
	assert.True(t, rules[2].IsEqualsTo(r3))
}

func TestFrequencyParamsRulesJsonConverter(t *testing.T) {
	t.Run("TestFrequencyParamsRulesJsonConverter_failed", func(t *testing.T) {
		properties, err := FrequencyParamsRulesJsonConverter([]byte{'s', 'r', 'c'})
		assert.True(t, properties == nil)
		assert.True(t, err != nil)
	})

	t.Run("TestFrequencyParamsRulesJsonConverter_Succeed", func(t *testing.T) {
		// Prepare test data
		f, err := os.Open("../../tests/testdata/extension/helper/FreqParamsRule.json")
		defer f.Close()
		if err != nil {
			t.Errorf("The rules file is not existed, err:%+v.", err)
		}
		src, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("Fail to read file, err: %+v.", err)
		}

		properties, err := FrequencyParamsRulesJsonConverter(src)
		rules := properties.([]*freq_params_traffic.Rule)
		assert.True(t, err == nil)
		assert.True(t, len(rules) == 4)
		for _, r := range rules {
			fmt.Println(r)
		}
		assert.True(t, strings.Contains(rules[0].String(), "Resource:abc, MetricType:Concurrency, Behavior:Reject, ParamIndex:0, Threshold:1000.000000, MaxQueueingTimeMs:1, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:map[{ValKind:KindInt ValStr:1000}:10001 {ValKind:KindString ValStr:ximu}:10002 {ValKind:KindBool ValStr:true}:10003]}"))
		assert.True(t, strings.Contains(rules[1].String(), "Resource:abc, MetricType:Concurrency, Behavior:Throttling, ParamIndex:1, Threshold:2000.000000, MaxQueueingTimeMs:2, BurstCount:20, DurationInSec:2, ParamsMaxCapacity:20000, SpecificItems:map[{ValKind:KindInt ValStr:1000}:20001 {ValKind:KindString ValStr:ximu}:20002 {ValKind:KindBool ValStr:true}:20003]}"))
		assert.True(t, strings.Contains(rules[2].String(), "Resource:abc, MetricType:QPS, Behavior:Reject, ParamIndex:2, Threshold:3000.000000, MaxQueueingTimeMs:3, BurstCount:30, DurationInSec:3, ParamsMaxCapacity:30000, SpecificItems:map[{ValKind:KindInt ValStr:1000}:30001 {ValKind:KindString ValStr:ximu}:30002 {ValKind:KindBool ValStr:true}:30003]}"))
		assert.True(t, strings.Contains(rules[3].String(), "Resource:abc, MetricType:QPS, Behavior:Throttling, ParamIndex:3, Threshold:4000.000000, MaxQueueingTimeMs:4, BurstCount:40, DurationInSec:4, ParamsMaxCapacity:40000, SpecificItems:map[{ValKind:KindInt ValStr:1000}:40001 {ValKind:KindString ValStr:ximu}:40002 {ValKind:KindBool ValStr:true}:40003]}"))
	})
}

func TestFrequencyParamsRulesUpdater(t *testing.T) {
	// Prepare test data
	m := make(map[freq_params_traffic.SpecificValue]int64)
	m[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindString,
		ValStr:  "sss",
	}] = 1
	m[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r1 := &freq_params_traffic.Rule{
		Id:                "1",
		Resource:          "abc",
		MetricType:        freq_params_traffic.Concurrency,
		Behavior:          freq_params_traffic.Reject,
		ParamIndex:        0,
		Threshold:         100,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     m,
	}

	m2 := make(map[freq_params_traffic.SpecificValue]int64)
	m2[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindString,
		ValStr:  "sss",
	}] = 1
	m2[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r2 := &freq_params_traffic.Rule{
		Id:                "2",
		Resource:          "abc",
		MetricType:        freq_params_traffic.QPS,
		Behavior:          freq_params_traffic.Throttling,
		ParamIndex:        1,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     m2,
	}

	m3 := make(map[freq_params_traffic.SpecificValue]int64)
	m3[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindString,
		ValStr:  "sss",
	}] = 1
	m3[freq_params_traffic.SpecificValue{
		ValKind: freq_params_traffic.KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r3 := &freq_params_traffic.Rule{
		Id:                "3",
		Resource:          "abc",
		MetricType:        freq_params_traffic.Concurrency,
		Behavior:          freq_params_traffic.Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     m3,
	}

	r4 := &freq_params_traffic.Rule{
		Id:                "4",
		Resource:          "abc",
		MetricType:        freq_params_traffic.Concurrency,
		Behavior:          freq_params_traffic.Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     2,
		SpecificItems:     m3,
	}

	err := FrequencyParamsRulesUpdater([]*freq_params_traffic.Rule{r1, r2, r3, r4})
	assert.True(t, err == nil)

	rules := freq_params_traffic.GetRules("abc")
	assert.True(t, rules[0].Equals(r1))
	assert.True(t, rules[1].Equals(r2))
	assert.True(t, rules[2].Equals(r3))
	assert.True(t, rules[3].Equals(r4))
}
