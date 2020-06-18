package hotspot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseSpecificItems(t *testing.T) {
	t.Run("Test_parseSpecificItems", func(t *testing.T) {
		source := make(map[SpecificValue]int64)
		s1 := SpecificValue{
			ValKind: KindInt,
			ValStr:  "10010",
		}
		s2 := SpecificValue{
			ValKind: KindInt,
			ValStr:  "10010aaa",
		}
		s3 := SpecificValue{
			ValKind: KindString,
			ValStr:  "test-string",
		}
		s4 := SpecificValue{
			ValKind: KindBool,
			ValStr:  "true",
		}
		s5 := SpecificValue{
			ValKind: KindFloat64,
			ValStr:  "1.234",
		}
		s6 := SpecificValue{
			ValKind: KindFloat64,
			ValStr:  "1.2345678",
		}
		source[s1] = 100
		source[s2] = 100
		source[s3] = 100
		source[s4] = 100
		source[s5] = 100
		source[s6] = 100

		got := parseSpecificItems(source)
		assert.True(t, len(got) == 5)
		assert.True(t, got[10010] == 100)
		assert.True(t, got[true] == 100)
		assert.True(t, got[1.234] == 100)
		assert.True(t, got[1.23400] == 100)
		assert.True(t, got["test-string"] == 100)
		assert.True(t, got[1.23457] == 100)
	})
}

func TestMetricType_String(t *testing.T) {
	t.Run("TestMetricType_String", func(t *testing.T) {
		assert.True(t, fmt.Sprintf("%+v", Concurrency) == "Concurrency")

	})
}

func Test_Rule_String(t *testing.T) {
	t.Run("Test_Rule_String_Normal", func(t *testing.T) {
		m := make(map[SpecificValue]int64)
		m[SpecificValue{
			ValKind: KindString,
			ValStr:  "sss",
		}] = 1
		m[SpecificValue{
			ValKind: KindFloat64,
			ValStr:  "1.123",
		}] = 3
		r := &Rule{
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
			SpecificItems:     m,
		}
		fmt.Println(fmt.Sprintf("%+v", []*Rule{r}))
		assert.True(t, fmt.Sprintf("%+v", []*Rule{r}) == "[{Id:abc, Resource:abc, MetricType:Concurrency, ControlBehavior:Reject, ParamIndex:0, Threshold:110.000000, MaxQueueingTimeMs:5, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:map[{ValKind:KindString ValStr:sss}:1 {ValKind:KindFloat64 ValStr:1.123}:3]}]")
	})
}

func Test_Rule_Equals(t *testing.T) {
	t.Run("Test_Rule_Equals", func(t *testing.T) {
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
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
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
			Id:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 5,
			BurstCount:        10,
			DurationInSec:     1,
			ParamsMaxCapacity: 10000,
			SpecificItems:     m2,
		}
		assert.True(t, r1.Equals(r2))
	})
}
