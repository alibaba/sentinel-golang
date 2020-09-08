package flow

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndRemoveTrafficShapingGenerator(t *testing.T) {
	tsc := &TrafficShapingController{}

	err := SetTrafficShapingGenerator(Direct, Reject, func(_ *Rule) *TrafficShapingController {
		return tsc
	})
	assert.Error(t, err, "default control behaviors are not allowed to be modified")
	err = RemoveTrafficShapingGenerator(Direct, Reject)
	assert.Error(t, err, "default control behaviors are not allowed to be removed")

	err = SetTrafficShapingGenerator(TokenCalculateStrategy(111), ControlBehavior(112), func(_ *Rule) *TrafficShapingController {
		return tsc
	})
	assert.NoError(t, err)

	resource := "test-customized-tc"
	_, err = LoadRules([]*Rule{
		{
			ID:                     10,
			Count:                  20,
			MetricType:             QPS,
			Resource:               resource,
			TokenCalculateStrategy: TokenCalculateStrategy(111),
			ControlBehavior:        ControlBehavior(112),
		},
	})

	cs := trafficControllerGenKey{
		tokenCalculateStrategy: TokenCalculateStrategy(111),
		controlBehavior:        ControlBehavior(112),
	}
	assert.NoError(t, err)
	assert.Contains(t, tcGenFuncMap, cs)
	assert.NotZero(t, len(tcMap[resource]))
	assert.Equal(t, tsc, tcMap[resource][0])

	err = RemoveTrafficShapingGenerator(TokenCalculateStrategy(111), ControlBehavior(112))
	assert.NoError(t, err)
	assert.NotContains(t, tcGenFuncMap, cs)

	_, _ = LoadRules([]*Rule{})
}

func TestIsValidFlowRule(t *testing.T) {
	badRule1 := &Rule{ID: 1, Count: 1, MetricType: QPS, Resource: ""}
	badRule2 := &Rule{ID: 1, Count: -1.9, MetricType: QPS, Resource: "test"}
	badRule3 := &Rule{Count: 5, MetricType: QPS, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Reject}
	goodRule1 := &Rule{Count: 10, MetricType: QPS, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Throttling, WarmUpPeriodSec: 10}

	assert.Error(t, IsValidRule(badRule1))
	assert.Error(t, IsValidRule(badRule2))
	assert.Error(t, IsValidRule(badRule3))
	assert.NoError(t, IsValidRule(goodRule1))
}

func TestGetRules(t *testing.T) {
	t.Run("GetRules", func(t *testing.T) {
		if err := ClearRules(); err != nil {
			t.Fatal(err)
		}
		r1 := &Rule{
			ID:                     0,
			Resource:               "abc1",
			MetricType:             0,
			Count:                  0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		r2 := &Rule{
			ID:                     1,
			Resource:               "abc2",
			MetricType:             0,
			Count:                  0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		if _, err := LoadRules([]*Rule{r1, r2}); err != nil {
			t.Fatal(err)
		}

		rs1 := GetRules()
		if rs1[0].Resource == "abc1" {
			assert.True(t, &rs1[0] != r1)
			assert.True(t, &rs1[1] != r2)
			assert.True(t, reflect.DeepEqual(&rs1[0], r1))
			assert.True(t, reflect.DeepEqual(&rs1[1], r2))
		} else {
			assert.True(t, &rs1[0] != r2)
			assert.True(t, &rs1[1] != r1)
			assert.True(t, reflect.DeepEqual(&rs1[0], r2))
			assert.True(t, reflect.DeepEqual(&rs1[1], r1))
		}
		if err := ClearRules(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("getRules", func(t *testing.T) {
		r1 := &Rule{
			ID:                     0,
			Resource:               "abc1",
			MetricType:             0,
			Count:                  0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		r2 := &Rule{
			ID:                     1,
			Resource:               "abc2",
			MetricType:             0,
			Count:                  0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		if _, err := LoadRules([]*Rule{r1, r2}); err != nil {
			t.Fatal(err)
		}
		rs2 := getRules()
		if rs2[0].Resource == "abc1" {
			assert.True(t, rs2[0] == r1)
			assert.True(t, rs2[1] == r2)
			assert.True(t, reflect.DeepEqual(rs2[0], r1))
			assert.True(t, reflect.DeepEqual(rs2[1], r2))
		} else {
			assert.True(t, rs2[0] == r2)
			assert.True(t, rs2[1] == r1)
			assert.True(t, reflect.DeepEqual(rs2[0], r2))
			assert.True(t, reflect.DeepEqual(rs2[1], r1))
		}
		if err := ClearRules(); err != nil {
			t.Fatal(err)
		}
	})
}
