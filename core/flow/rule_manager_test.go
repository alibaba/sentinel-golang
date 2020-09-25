package flow

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/stat"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/stretchr/testify/assert"
)

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

	_, _ = LoadRules([]*Rule{})
}

func TestIsValidFlowRule(t *testing.T) {
	badRule1 := &Rule{Threshold: 1, Resource: ""}
	badRule2 := &Rule{Threshold: -1.9, Resource: "test"}
	badRule3 := &Rule{Threshold: 5, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Reject}
	goodRule1 := &Rule{Threshold: 10, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Throttling, WarmUpPeriodSec: 10, MaxQueueingTimeMs: 10}
	badRule4 := &Rule{Threshold: 5, Resource: "test", TokenCalculateStrategy: WarmUp, ControlBehavior: Reject, StatIntervalInMs: 6000000}

	assert.Error(t, IsValidRule(badRule1))
	assert.Error(t, IsValidRule(badRule2))
	assert.Error(t, IsValidRule(badRule3))
	assert.NoError(t, IsValidRule(goodRule1))
	assert.Error(t, IsValidRule(badRule4))
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

func Test_buildRulesOfRes(t *testing.T) {
	t.Run("Test_buildRulesOfRes_no_reuse_stat", func(t *testing.T) {
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
		tcs := buildRulesOfRes("abc1", []*Rule{r1, r2})
		assert.True(t, len(tcs) == 2)
		assert.True(t, tcs[0].BoundRule() == r1)
		assert.True(t, tcs[1].BoundRule() == r2)
		assert.True(t, reflect.DeepEqual(tcs[0].BoundRule(), r1))
		assert.True(t, reflect.DeepEqual(tcs[1].BoundRule(), r2))
	})

	t.Run("Test_buildRulesOfRes_reuse_stat", func(t *testing.T) {
		// reuse
		r1 := &Rule{
			Resource:               "abc1",
			Threshold:              100,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       1000,
		}
		// reuse
		r2 := &Rule{
			Resource:               "abc1",
			Threshold:              200,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       2000,
		}
		// reuse
		r3 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       5000,
		}
		// independent statistic
		r4 := &Rule{
			Resource:               "abc1",
			Threshold:              400,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       50000,
		}

		s1, err := generateStatFor(r1)
		if err != nil {
			t.Fatal(err)
		}
		fakeTc1 := &TrafficShapingController{flowCalculator: nil, flowChecker: nil, rule: r1, boundStat: *s1}
		s2, err := generateStatFor(r2)
		if err != nil {
			t.Fatal(err)
		}
		fakeTc2 := &TrafficShapingController{flowCalculator: nil, flowChecker: nil, rule: r2, boundStat: *s2}
		s3, err := generateStatFor(r3)
		if err != nil {
			t.Fatal(err)
		}
		fakeTc3 := &TrafficShapingController{flowCalculator: nil, flowChecker: nil, rule: r3, boundStat: *s3}
		s4, err := generateStatFor(r4)
		if err != nil {
			t.Fatal(err)
		}
		fakeTc4 := &TrafficShapingController{flowCalculator: nil, flowChecker: nil, rule: r4, boundStat: *s4}
		tcMap["abc1"] = []*TrafficShapingController{fakeTc1, fakeTc2, fakeTc3, fakeTc4}
		assert.True(t, len(tcMap["abc1"]) == 4)
		stat1 := tcMap["abc1"][0].boundStat
		stat2 := tcMap["abc1"][1].boundStat
		oldTc3 := tcMap["abc1"][2]
		assert.True(t, tcMap["abc1"][3].boundStat.writeOnlyMetric != nil)
		assert.True(t, !tcMap["abc1"][3].boundStat.reuseResourceStat)
		stat4 := tcMap["abc1"][3].boundStat

		// reuse stat
		r12 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			StatIntervalInMs:       1000,
		}
		// not reusable, generate from resource's global statistic
		r22 := &Rule{
			Resource:               "abc1",
			Threshold:              400,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Throttling,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       10000,
		}
		// equals
		r32 := &Rule{
			Resource:               "abc1",
			Threshold:              300,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       5000,
		}
		// reuse independent stat
		r42 := &Rule{
			Resource:               "abc1",
			Threshold:              4000,
			RelationStrategy:       CurrentResource,
			TokenCalculateStrategy: Direct,
			ControlBehavior:        Reject,
			MaxQueueingTimeMs:      10,
			StatIntervalInMs:       50000,
		}
		tcs := buildRulesOfRes("abc1", []*Rule{r12, r22, r32, r42})
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
		assert.True(t, tcs[2] == oldTc3)
		assert.True(t, tcs[3].boundStat == stat4)
	})
}
