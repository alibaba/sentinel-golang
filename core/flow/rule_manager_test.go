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

package flow

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/stat"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/stretchr/testify/assert"
)

func clearData() {
	tcMap = make(TrafficControllerMap)
	currentRules = make(map[string][]*Rule, 0)
}
func TestSetAndRemoveTrafficShapingGenerator(t *testing.T) {
	tsc := &TrafficShapingController{}

	err := SetTrafficShapingGenerator(Direct, Reject, func(_ *Rule, _ *standaloneStatistic) (*TrafficShapingController, error) {
		return tsc, nil
	})
	assert.Error(t, err, "default control behaviors are not allowed to be modified")
	err = RemoveTrafficShapingGenerator(Direct, Reject)
	assert.Error(t, err, "default control behaviors are not allowed to be removed")

	err = SetTrafficShapingGenerator(TokenCalculateStrategy(111), ControlBehavior(112), func(_ *Rule, _ *standaloneStatistic) (*TrafficShapingController, error) {
		return tsc, nil
	})
	assert.NoError(t, err)

	resource := "test-customized-tc"
	_, err = LoadRules([]*Rule{
		{
			Threshold:              20,
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

	clearData()
}

func TestIsValidFlowRule(t *testing.T) {
	badRule1 := &Rule{Threshold: 1, Resource: ""}
	badRule2 := &Rule{Threshold: -1.9, Resource: "test"}
	badRule3 := &Rule{Threshold: 5, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Reject}
	badRule4 := &Rule{Threshold: 5, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Reject, StatIntervalInMs: 6000000}

	goodRule1 := &Rule{Threshold: 10, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Throttling, WarmUpPeriodSec: 10, MaxQueueingTimeMs: 10, StatIntervalInMs: 1000}
	goodRule2 := &Rule{Threshold: 10, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Throttling, WarmUpPeriodSec: 10, MaxQueueingTimeMs: 0, StatIntervalInMs: 1000}

	assert.Error(t, IsValidRule(badRule1))
	assert.Error(t, IsValidRule(badRule2))
	assert.Error(t, IsValidRule(badRule3))
	assert.Error(t, IsValidRule(badRule4))

	assert.NoError(t, IsValidRule(goodRule1))
	assert.NoError(t, IsValidRule(goodRule2))
}

func TestGetRules(t *testing.T) {
	t.Run("GetRules", func(t *testing.T) {
		if err := ClearRules(); err != nil {
			t.Fatal(err)
		}
		r1 := &Rule{
			Resource:               "abc1",
			Threshold:              0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		r2 := &Rule{
			Resource:               "abc2",
			Threshold:              0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       1000,
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
		clearData()
	})

	t.Run("getRules", func(t *testing.T) {
		r1 := &Rule{
			Resource:               "abc1",
			Threshold:              0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      0,
		}
		r2 := &Rule{
			Resource:               "abc2",
			Threshold:              0,
			RelationStrategy:       0,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			RefResource:            "",
			WarmUpPeriodSec:        0,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       1000,
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
		assert.True(t, len(tcMap["abc2"]) == 1 && !tcMap["abc2"][0].boundStat.reuseResourceStat)
		assert.True(t, reflect.DeepEqual(tcMap["abc2"][0].boundStat.readOnlyMetric, nopStat.readOnlyMetric))
		assert.True(t, reflect.DeepEqual(tcMap["abc2"][0].boundStat.writeOnlyMetric, nopStat.writeOnlyMetric))
		clearData()
	})
}

func Test_generateStatFor(t *testing.T) {
	t.Run("generateStatFor_reuse_default_metric_stat", func(t *testing.T) {
		r1 := &Rule{
			Resource:               "abc",
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       0,
			Threshold:              100,
			RelationStrategy:       CurrentResource,
		}
		// global: 10000ms, 20 sample, bucketLen: 500ms
		// metric: 1000ms,  2 sample,  bucketLen: 500ms
		boundStat, err := generateStatFor(r1)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, boundStat.reuseResourceStat && boundStat.writeOnlyMetric == nil)
		ps, succ := boundStat.readOnlyMetric.(*sbase.SlidingWindowMetric)
		assert.True(t, succ)

		resNode := stat.GetResourceNode("abc")
		assert.True(t, reflect.DeepEqual(ps, resNode.DefaultMetric()))
	})

	t.Run("generateStatFor_reuse_global_stat", func(t *testing.T) {
		// global: 10000ms, 20 sample, bucketLen: 500ms
		// metric: 1000ms,  2 sample,  bucketLen: 500ms
		r1 := &Rule{
			Resource:               "abc",
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       5000,
			Threshold:              100,
			RelationStrategy:       CurrentResource,
			RefResource:            "",
		}
		// global: 10000ms, 20 sample, bucketLen: 500ms
		// metric: 1000ms,  2 sample,  bucketLen: 500ms
		boundStat, err := generateStatFor(r1)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, boundStat.reuseResourceStat && boundStat.writeOnlyMetric == nil)
		ps, succ := boundStat.readOnlyMetric.(*sbase.SlidingWindowMetric)
		assert.True(t, succ)
		resNode := stat.GetResourceNode("abc")
		assert.True(t, !reflect.DeepEqual(ps, resNode.DefaultMetric()))
	})

	t.Run("generateStatFor_standalone_stat", func(t *testing.T) {
		// global: 10000ms, 20 sample, bucketLen: 500ms
		// metric: 1000ms,  2 sample,  bucketLen: 500ms
		r1 := &Rule{
			Resource:               "abc",
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       50000,
			Threshold:              100,
			RelationStrategy:       CurrentResource,
		}
		// global: 10000ms, 20 sample, bucketLen: 500ms
		// metric: 1000ms,  2 sample,  bucketLen: 500ms
		boundStat, err := generateStatFor(r1)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, boundStat.reuseResourceStat == false && boundStat.writeOnlyMetric != nil)
	})
}

func Test_buildResourceTrafficShapingController(t *testing.T) {
	t.Run("Test_buildResourceTrafficShapingController_no_reuse_stat", func(t *testing.T) {
		r1 := &Rule{
			Resource:               "abc1",
			Threshold:              100,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
		}
		r2 := &Rule{
			Resource:               "abc1",
			Threshold:              200,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			MaxQueueingTimeMs:      10,
		}
		assert.True(t, len(tcMap["abc1"]) == 0)
		tcs := buildResourceTrafficShapingController("abc1", []*Rule{r1, r2}, tcMap["abc1"])
		assert.True(t, len(tcs) == 2)
		assert.True(t, tcs[0].BoundRule() == r1)
		assert.True(t, tcs[1].BoundRule() == r2)
		assert.True(t, reflect.DeepEqual(tcs[0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(tcs[1].BoundRule(), r2))

		clearData()
	})

	t.Run("Test_buildResourceTrafficShapingController_reuse_stat", func(t *testing.T) {
		// use nop statistics because of no need statistics
		r0 := &Rule{
			Resource:               "abc1",
			Threshold:              100,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			StatIntervalInMs:       1000,
		}
		// reuse resource node default leap array
		r1 := &Rule{
			Resource:               "abc1",
			Threshold:              100,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       1000,
		}
		// reuse resource node default leap array
		r2 := &Rule{
			Resource:               "abc1",
			Threshold:              200,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       2000,
		}
		// reuse resource node default leap array
		r3 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       5000,
		}
		// use independent leap array because of too big interval
		r4 := &Rule{
			Resource:               "abc1",
			Threshold:              400,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       50000,
		}

		s0, err := generateStatFor(r0)
		assert.Empty(t, err)
		fakeTc0, err := NewTrafficShapingController(r0, s0)
		assert.Empty(t, err)
		stat0 := fakeTc0.boundStat
		assert.True(t, reflect.DeepEqual(stat0, *nopStat))
		assert.True(t, stat0.reuseResourceStat == false)
		assert.True(t, stat0.readOnlyMetric != nil)
		assert.True(t, stat0.writeOnlyMetric != nil)

		s1, err := generateStatFor(r1)
		assert.Empty(t, err)
		fakeTc1, err := NewTrafficShapingController(r1, s1)
		assert.Empty(t, err)
		stat1 := fakeTc1.boundStat
		assert.True(t, !reflect.DeepEqual(stat1, stat0))
		assert.True(t, stat1.reuseResourceStat == true)
		assert.True(t, stat1.readOnlyMetric != nil)
		assert.True(t, stat1.writeOnlyMetric == nil)

		s2, err := generateStatFor(r2)
		assert.Empty(t, err)
		fakeTc2, err := NewTrafficShapingController(r2, s2)
		assert.Empty(t, err)
		stat2 := fakeTc2.boundStat
		assert.True(t, !reflect.DeepEqual(stat2, stat0))
		assert.True(t, stat2.reuseResourceStat == true)
		assert.True(t, stat2.readOnlyMetric != nil)
		assert.True(t, stat2.writeOnlyMetric == nil)

		s3, err := generateStatFor(r3)
		assert.Empty(t, err)
		fakeTc3, err := NewTrafficShapingController(r3, s3)
		assert.Empty(t, err)
		stat3 := fakeTc3.boundStat
		assert.True(t, !reflect.DeepEqual(stat3, stat0))
		assert.True(t, stat3.reuseResourceStat == true)
		assert.True(t, stat3.readOnlyMetric != nil)
		assert.True(t, stat3.writeOnlyMetric == nil)

		s4, err := generateStatFor(r4)
		assert.Empty(t, err)
		fakeTc4, err := NewTrafficShapingController(r4, s4)
		assert.Empty(t, err)
		stat4 := fakeTc4.boundStat
		assert.True(t, !reflect.DeepEqual(stat4, stat0))
		assert.True(t, stat4.reuseResourceStat == false)
		assert.True(t, stat4.readOnlyMetric != nil)
		assert.True(t, stat4.writeOnlyMetric != nil)

		tcMap["abc1"] = []*TrafficShapingController{fakeTc0, fakeTc1, fakeTc2, fakeTc3, fakeTc4}
		assert.True(t, len(tcMap["abc1"]) == 5)
		// reuse stat with rule 1
		r12 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       1000,
		}
		// can't reuse stat with rule 2, generate from resource's global statistic
		r22 := &Rule{
			Resource:               "abc1",
			Threshold:              400,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       10000,
		}
		// equals with rule 3
		r32 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       5000,
		}
		// reuse independent stat with rule 4
		r42 := &Rule{
			Resource:               "abc1",
			Threshold:              4000,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       50000,
		}

		tcs := buildResourceTrafficShapingController("abc1", []*Rule{r12, r22, r32, r42}, tcMap["abc1"])
		assert.True(t, len(tcs) == 4)

		assert.True(t, tcs[0].BoundRule() == r12)
		assert.True(t, tcs[1].BoundRule() == r22)
		assert.True(t, tcs[2].BoundRule() == r3)
		assert.True(t, tcs[3].BoundRule() == r42)

		assert.True(t, reflect.DeepEqual(tcs[0].BoundRule(), r12))
		assert.True(t, reflect.DeepEqual(tcs[1].BoundRule(), r22))
		assert.True(t, reflect.DeepEqual(tcs[2].BoundRule(), r32) && reflect.DeepEqual(tcs[2].BoundRule(), r3))
		assert.True(t, reflect.DeepEqual(tcs[3].BoundRule(), r42))

		assert.True(t, tcs[0].boundStat == stat1)
		assert.True(t, tcs[1].boundStat != stat2)
		assert.True(t, tcs[2] == fakeTc3)
		assert.True(t, tcs[3].boundStat == stat4)

		clearData()
	})
}

func TestLoadRules(t *testing.T) {
	t.Run("loadSameRules", func(t *testing.T) {
		_, err := LoadRules([]*Rule{
			{
				Resource:               "some-test",
				Threshold:              10,
				TokenCalculateStrategy: Direct,
				ControlBehavior:        Reject,
			},
		})
		assert.Nil(t, err)
		ok, err := LoadRules([]*Rule{
			{
				Resource:               "some-test",
				Threshold:              10,
				TokenCalculateStrategy: Direct,
				ControlBehavior:        Reject,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)

		clearData()
	})
}

func TestIsValidRule(t *testing.T) {
	rule1 := &Rule{
		Resource:               "hello0",
		TokenCalculateStrategy: MemoryAdaptive,
		ControlBehavior:        Reject,
		StatIntervalInMs:       10,
		LowMemUsageThreshold:   2,
		HighMemUsageThreshold:  1,
		MemLowWaterMarkBytes:   1,
		MemHighWaterMarkBytes:  2,
	}
	assert.Nil(t, IsValidRule(rule1))

	rule1.LowMemUsageThreshold = 9
	rule1.HighMemUsageThreshold = 9
	assert.NotNil(t, IsValidRule(rule1))
	rule1.LowMemUsageThreshold = 10
	assert.Nil(t, IsValidRule(rule1))

	rule1.MemLowWaterMarkBytes = 0
	assert.NotNil(t, IsValidRule(rule1))
	rule1.MemLowWaterMarkBytes = 100 * 1024 * 1024
	rule1.MemHighWaterMarkBytes = 300 * 1024 * 1024
	assert.Nil(t, IsValidRule(rule1))

	rule1.MemHighWaterMarkBytes = 0
	assert.NotNil(t, IsValidRule(rule1))
	rule1.MemHighWaterMarkBytes = 300 * 1024 * 1024
	assert.Nil(t, IsValidRule(rule1))

	rule1.MemLowWaterMarkBytes = 100 * 1024 * 1024
	rule1.MemHighWaterMarkBytes = 30 * 1024 * 1024
	assert.NotNil(t, IsValidRule(rule1))
	rule1.MemHighWaterMarkBytes = 300 * 1024 * 1024
	assert.Nil(t, IsValidRule(rule1))
}

func TestLoadRulesOfResource(t *testing.T) {
	r11 := &Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              10,
	}
	r12 := &Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              20,
	}
	r21 := &Rule{
		Resource:               "abc2",
		Threshold:              10,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}
	r22 := &Rule{
		Resource:               "abc2",
		Threshold:              20,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}

	succ, err := LoadRules([]*Rule{r11, r12, r21, r22})
	assert.True(t, succ && err == nil)

	t.Run("LoadRulesOfResource_empty_resource", func(t *testing.T) {
		succ, err = LoadRulesOfResource("", []*Rule{r11, r12})
		assert.True(t, !succ && err != nil)
	})

	t.Run("LoadRulesOfResource_cache_hit", func(t *testing.T) {
		r111 := *r11
		r122 := *r12
		succ, err = LoadRulesOfResource("abc1", []*Rule{&r111, &r122})
		assert.True(t, !succ && err == nil)
	})

	t.Run("LoadRulesOfResource_clear", func(t *testing.T) {
		succ, err = LoadRulesOfResource("abc1", []*Rule{})
		assert.True(t, succ && err == nil)
		assert.True(t, len(tcMap["abc1"]) == 0 && len(currentRules["abc1"]) == 0)
		assert.True(t, len(tcMap["abc2"]) == 2 && len(currentRules["abc2"]) == 2)
	})
	clearData()
}

func Test_onResourceRuleUpdate(t *testing.T) {
	r11 := Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              10,
	}
	r12 := Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              20,
	}
	r21 := Rule{
		Resource:               "abc2",
		Threshold:              10,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}
	r22 := Rule{
		Resource:               "abc2",
		Threshold:              20,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}

	succ, err := LoadRules([]*Rule{&r11, &r12, &r21, &r22})
	assert.True(t, succ && err == nil)

	t.Run("Test_onResourceRuleUpdate_normal", func(t *testing.T) {
		r111 := r11
		r111.Threshold = 100
		err = onResourceRuleUpdate("abc1", []*Rule{&r111})

		assert.True(t, len(tcMap["abc1"]) == 1)
		assert.True(t, len(currentRules["abc1"]) == 1)
		assert.True(t, tcMap["abc1"][0].rule == &r111)

		assert.True(t, len(tcMap["abc2"]) == 2)
		assert.True(t, len(currentRules["abc2"]) == 2)

		clearData()
	})
}

func TestClearRulesOfResource(t *testing.T) {
	r11 := Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              10,
	}
	r12 := Rule{
		Resource:               "abc1",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		Threshold:              20,
	}
	r21 := Rule{
		Resource:               "abc2",
		Threshold:              10,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}
	r22 := Rule{
		Resource:               "abc2",
		Threshold:              20,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
	}

	succ, err := LoadRules([]*Rule{&r11, &r12, &r21, &r22})
	assert.True(t, succ && err == nil)

	t.Run("TestClearRulesOfResource_normal", func(t *testing.T) {
		assert.True(t, ClearRulesOfResource("abc1") == nil)

		assert.True(t, len(tcMap["abc1"]) == 0)
		assert.True(t, len(currentRules["abc1"]) == 0)
		assert.True(t, len(tcMap["abc2"]) == 2)
		assert.True(t, len(currentRules["abc2"]) == 2)
		clearData()
	})
}
