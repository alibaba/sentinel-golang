package hotspot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseSpecificItems(t *testing.T) {
	t.Run("Test_parseSpecificItems", func(t *testing.T) {
		source := make([]SpecificValue, 6)
		s1 := SpecificValue{
			ValKind:   KindInt,
			ValStr:    "10010",
			Threshold: 100,
		}
		s2 := SpecificValue{
			ValKind:   KindInt,
			ValStr:    "10010aaa",
			Threshold: 100,
		}
		s3 := SpecificValue{
			ValKind:   KindString,
			ValStr:    "test-string",
			Threshold: 100,
		}
		s4 := SpecificValue{
			ValKind:   KindBool,
			ValStr:    "true",
			Threshold: 100,
		}
		s5 := SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.234",
			Threshold: 100,
		}
		s6 := SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.2345678",
			Threshold: 100,
		}
		source[0] = s1
		source[1] = s2
		source[2] = s3
		source[3] = s4
		source[4] = s5
		source[5] = s6

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
		m := make([]SpecificValue, 2)
		m[0] = SpecificValue{
			ValKind:   KindString,
			ValStr:    "sss",
			Threshold: 1,
		}
		m[1] = SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.123",
			Threshold: 3,
		}
		r := &Rule{
			ID:                "abc",
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
		assert.True(t, fmt.Sprintf("%+v", []*Rule{r}) == "[{Id:abc, Resource:abc, MetricType:Concurrency, ControlBehavior:Reject, ParamIndex:0, Threshold:110.000000, MaxQueueingTimeMs:5, BurstCount:10, DurationInSec:1, ParamsMaxCapacity:10000, SpecificItems:[{ValKind:KindString ValStr:sss Threshold:1} {ValKind:KindFloat64 ValStr:1.123 Threshold:3}]}]")
	})
}

func Test_Rule_Equals(t *testing.T) {
	t.Run("Test_Rule_Equals", func(t *testing.T) {
		m := make([]SpecificValue, 2)
		m[0] = SpecificValue{
			ValKind:   KindString,
			ValStr:    "sss",
			Threshold: 1,
		}
		m[1] = SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.123",
			Threshold: 3,
		}
		r1 := &Rule{
			ID:                "abc",
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

		m2 := make([]SpecificValue, 2)
		m2[0] = SpecificValue{
			ValKind:   KindString,
			ValStr:    "sss",
			Threshold: 1,
		}
		m2[1] = SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.123",
			Threshold: 3,
		}
		r2 := &Rule{
			ID:                "abc",
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
