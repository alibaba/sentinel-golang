package hotspot

import (
	"fmt"
	"math"
	"testing"

	"github.com/alibaba/sentinel-golang/core/hotspot/cache"
	"github.com/stretchr/testify/assert"
)

func Test_tcGenFuncMap(t *testing.T) {
	t.Run("Test_tcGenFuncMap_withoutMetric", func(t *testing.T) {
		m := make(map[SpecificValue]int64)
		m[SpecificValue{
			ValKind: KindInt,
			ValStr:  "100",
		}] = 100

		parsedM := make(map[interface{}]int64)
		parsedM[100] = 100
		r1 := &Rule{
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     m,
		}
		generator, supported := tcGenFuncMap[r1.ControlBehavior]
		assert.True(t, supported && generator != nil)
		tc := generator(r1, nil)
		assert.True(t, tc.BoundMetric() != nil && tc.BoundRule() == r1 && tc.BoundParamIndex() == 0)
		rejectTC := tc.(*rejectTrafficShapingController)
		assert.True(t, rejectTC != nil)
		assert.True(t, rejectTC.res == r1.Resource && rejectTC.metricType == r1.MetricType && rejectTC.paramIndex == r1.ParamIndex && rejectTC.burstCount == r1.BurstCount)
		assert.True(t, rejectTC.threshold == r1.Threshold && rejectTC.durationInSec == r1.DurationInSec)
	})

	t.Run("Test_tcGenFuncMap_withMetric", func(t *testing.T) {
		m := make(map[SpecificValue]int64)
		m[SpecificValue{
			ValKind: KindInt,
			ValStr:  "100",
		}] = 100

		parsedM := make(map[interface{}]int64)
		parsedM[100] = 100
		r1 := &Rule{
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     m,
		}
		generator, supported := tcGenFuncMap[r1.ControlBehavior]
		assert.True(t, supported && generator != nil)

		size := int(math.Min(float64(ParamsMaxCapacity), float64(ParamsCapacityBase*r1.DurationInSec)))
		if size <= 0 {
			size = ParamsMaxCapacity
		}
		metric := &ParamsMetric{
			RuleTimeCounter:    cache.NewLRUCacheMap(size),
			RuleTokenCounter:   cache.NewLRUCacheMap(size),
			ConcurrencyCounter: cache.NewLRUCacheMap(ConcurrencyMaxCount),
		}

		tc := generator(r1, metric)
		assert.True(t, tc.BoundMetric() != nil && tc.BoundRule() == r1 && tc.BoundParamIndex() == 0)
		rejectTC := tc.(*rejectTrafficShapingController)
		assert.True(t, rejectTC != nil)
		assert.True(t, rejectTC.metric == metric)
		assert.True(t, rejectTC.res == r1.Resource && rejectTC.metricType == r1.MetricType && rejectTC.paramIndex == r1.ParamIndex && rejectTC.burstCount == r1.BurstCount)
		assert.True(t, rejectTC.threshold == r1.Threshold && rejectTC.durationInSec == r1.DurationInSec)

	})
}

func Test_IsValidRule(t *testing.T) {
	t.Run("Test_IsValidRule", func(t *testing.T) {
		m := make(map[SpecificValue]int64)
		m[SpecificValue{
			ValKind: KindInt,
			ValStr:  "100",
		}] = 100

		parsedM := make(map[interface{}]int64)
		parsedM[100] = 100
		r1 := &Rule{
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     m,
		}
		assert.True(t, IsValidRule(r1) == nil)
	})

	t.Run("Test_InValidRule", func(t *testing.T) {
		m := make(map[SpecificValue]int64)
		m[SpecificValue{
			ValKind: KindInt,
			ValStr:  "100",
		}] = 100

		parsedM := make(map[interface{}]int64)
		parsedM[100] = 100
		r1 := &Rule{
			Id:                "",
			Resource:          "",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     m,
		}
		assert.True(t, IsValidRule(r1) != nil)
	})
}

func Test_buildTcMap(t *testing.T) {
	m := make(map[SpecificValue]int64)
	m[SpecificValue{
		ValKind: KindString,
		ValStr:  "sss",
	}] = 1
	m[SpecificValue{
		ValKind: KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r1 := &Rule{
		Id:                "1",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Reject,
		ParamIndex:        0,
		Threshold:         100,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     m,
	}

	m2 := make(map[SpecificValue]int64)
	m2[SpecificValue{
		ValKind: KindString,
		ValStr:  "sss",
	}] = 1
	m2[SpecificValue{
		ValKind: KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r2 := &Rule{
		Id:                "2",
		Resource:          "abc",
		MetricType:        QPS,
		ControlBehavior:   Throttling,
		ParamIndex:        1,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     m2,
	}

	m3 := make(map[SpecificValue]int64)
	m3[SpecificValue{
		ValKind: KindString,
		ValStr:  "sss",
	}] = 1
	m3[SpecificValue{
		ValKind: KindFloat64,
		ValStr:  "1.123",
	}] = 3
	r3 := &Rule{
		Id:                "3",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     m3,
	}

	r4 := &Rule{
		Id:                "4",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     2,
		SpecificItems:     m3,
	}

	updated, err := LoadRules([]*Rule{r1, r2, r3, r4})
	if !updated || err != nil {
		t.Errorf("Fail to prepare data, err: %+v", err)
	}
	assert.True(t, len(tcMap["abc"]) == 4)

	r21 := &Rule{
		Id:                "21",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Reject,
		ParamIndex:        0,
		Threshold:         100,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     m,
	}
	r22 := &Rule{
		Id:                "22",
		Resource:          "abc",
		MetricType:        QPS,
		ControlBehavior:   Throttling,
		ParamIndex:        1,
		Threshold:         101,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     m2,
	}
	r23 := &Rule{
		Id:                "23",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     12,
		SpecificItems:     m3,
	}

	oldTc1Ptr := tcMap["abc"][0]
	oldTc2Ptr := tcMap["abc"][1]
	oldTc3Ptr := tcMap["abc"][2]
	oldTc4Ptr := tcMap["abc"][3]
	oldTc1PtrAddr := fmt.Sprintf("%p", oldTc1Ptr)
	oldTc2PtrAddr := fmt.Sprintf("%p", oldTc2Ptr)
	oldTc3PtrAddr := fmt.Sprintf("%p", oldTc3Ptr)
	oldTc4PtrAddr := fmt.Sprintf("%p", oldTc4Ptr)
	fmt.Println(oldTc1PtrAddr)
	fmt.Println(oldTc2PtrAddr)
	fmt.Println(oldTc3PtrAddr)
	fmt.Println(oldTc4PtrAddr)
	oldTc2MetricPtrAddr := fmt.Sprintf("%p", tcMap["abc"][1].BoundMetric())
	fmt.Println("oldTc2MetricPtr:", oldTc2MetricPtrAddr)

	newTcMap := buildTcMap([]*Rule{r21, r22, r23})
	assert.True(t, len(newTcMap) == 1)
	abcTcs := newTcMap["abc"]
	assert.True(t, len(abcTcs) == 3)
	newTc1Ptr := abcTcs[0]
	newTc2Ptr := abcTcs[1]
	newTc3Ptr := abcTcs[2]
	newTc1PtrAddr := fmt.Sprintf("%p", newTc1Ptr)
	newTc2PtrAddr := fmt.Sprintf("%p", newTc2Ptr)
	newTc3PtrAddr := fmt.Sprintf("%p", newTc3Ptr)
	fmt.Println(newTc1PtrAddr)
	fmt.Println(newTc2PtrAddr)
	fmt.Println(newTc3PtrAddr)
	newTc2MetricPtrAddr := fmt.Sprintf("%p", newTc2Ptr.BoundMetric())
	fmt.Println("newTc2MetricPtrAddr:", newTc2MetricPtrAddr)
	assert.True(t, newTc1PtrAddr == oldTc1PtrAddr && newTc2MetricPtrAddr == oldTc2MetricPtrAddr)
	assert.True(t, abcTcs[0].BoundRule() == r1 && abcTcs[0] == oldTc1Ptr)
	assert.True(t, abcTcs[1].BoundMetric() == oldTc2Ptr.BoundMetric())

	tcMap = make(trafficControllerMap)
}
