package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRules(t *testing.T) {
	t.Run("EmptyRules", func(t *testing.T) {
		rules := GetRules()
		assert.Equal(t, 0, len(rules))
	})

	t.Run("GetUpdatedRules", func(t *testing.T) {
		defer func() { ruleMap = make(RuleMap, 0) }()

		r := map[MetricType][]*SystemRule{
			InboundQPS:  {&SystemRule{MetricType: InboundQPS, TriggerCount: 1}},
			Concurrency: {&SystemRule{MetricType: Concurrency, TriggerCount: 2}},
		}
		ruleMap = r
		rules := GetRules()
		assert.Equal(t, 2, len(rules))

		r[InboundQPS] = append(r[InboundQPS], &SystemRule{MetricType: InboundQPS, TriggerCount: 2})
		ruleMap = r
		rules = GetRules()
		assert.Equal(t, 3, len(rules))
	})
}

func TestLoadRules(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		err := LoadRules(nil)
		assert.NoError(t, err)
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		sRule := []*SystemRule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		err := LoadRules(sRule)
		assert.NoError(t, err)
	})
}

func TestOnRuleUpdate(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		err := onRuleUpdate(nil)
		assert.NoError(t, err)
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		defer func() { ruleMap = make(RuleMap, 0) }()

		sRule := []*SystemRule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		err := onRuleUpdate(sRule)
		assert.NoError(t, err)
		assert.Equal(t, len(sRule), len(ruleMap))
	})
}

func TestBuildRuleMap(t *testing.T) {
	t.Run("NilSystemRule", func(t *testing.T) {
		r := buildRuleMap(nil)
		assert.Equal(t, 0, len(r))
	})

	t.Run("InvalidSystemRule", func(t *testing.T) {
		sRule := []*SystemRule{
			{MetricType: InboundQPS, TriggerCount: -1},
		}
		r := buildRuleMap(sRule)
		assert.Equal(t, 0, len(r))
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		sRule := []*SystemRule{
			{MetricType: InboundQPS, TriggerCount: 1},
			{MetricType: Concurrency, TriggerCount: 2},
		}
		r := buildRuleMap(sRule)
		assert.Equal(t, len(sRule), len(r))
	})

	t.Run("MultiRuleOneTypeValidSystemRule", func(t *testing.T) {
		sRule := []*SystemRule{
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
		assert.EqualError(t, err, "nil SystemRule")
	})

	t.Run("NegativeThreshold", func(t *testing.T) {
		sRule := &SystemRule{MetricType: InboundQPS, TriggerCount: -1}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "negative threshold")
	})

	t.Run("InvalidMetricType", func(t *testing.T) {
		sRule := &SystemRule{MetricType: MetricTypeSize}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "invalid metric type")
	})

	t.Run("InvalidCPUUsage", func(t *testing.T) {
		sRule := &SystemRule{MetricType: CpuUsage, TriggerCount: 75}
		err := IsValidSystemRule(sRule)
		assert.EqualError(t, err, "invalid CPU usage, valid range is [0.0, 1.0]")
	})

	t.Run("ValidSystemRule", func(t *testing.T) {
		sRule := &SystemRule{MetricType: Load, TriggerCount: 12, Strategy: BBR}
		err := IsValidSystemRule(sRule)
		assert.NoError(t, err)
	})
}
