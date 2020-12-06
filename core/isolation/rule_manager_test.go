package isolation

import (
	"testing"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/stretchr/testify/assert"
)

func TestLoadRules(t *testing.T) {
	t.Run("TestLoadRules_1", func(t *testing.T) {
		logging.ResetGlobalLoggerLevel(logging.DebugLevel)
		r1 := &Rule{
			Resource:   "abc1",
			MetricType: Concurrency,
			Threshold:  100,
		}
		r2 := &Rule{
			Resource:   "abc2",
			MetricType: Concurrency,
			Threshold:  200,
		}
		r3 := &Rule{
			Resource:   "abc3",
			MetricType: MetricType(1),
			Threshold:  200,
		}
		_, err := LoadRules([]*Rule{r1, r2, r3})
		assert.True(t, err == nil)
		assert.True(t, len(RuleMap) == 2)
		assert.True(t, len(RuleMap["abc1"]) == 1)
		assert.True(t, RuleMap["abc1"][0] == r1)
		assert.True(t, len(RuleMap["abc2"]) == 1)
		assert.True(t, RuleMap["abc2"][0] == r2)

		err = ClearRules()
		assert.True(t, err == nil)
	})

	t.Run("loadSameRules", func(t *testing.T) {
		_, err := LoadRules([]*Rule{
			{
				Resource:   "abc1",
				MetricType: Concurrency,
				Threshold:  100,
			},
		})
		assert.Nil(t, err)
		ok, err := LoadRules([]*Rule{
			{
				Resource:   "abc1",
				MetricType: Concurrency,
				Threshold:  100,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestWhenUpdateRules(t *testing.T) {
	t.Run("WhenUpdateRules", func(t *testing.T) {
		_ = ClearRules()
		WhenUpdateRules(ruleUpdateForResetResourceHandler)
		_, err := LoadRules([]*Rule{
			{
				Resource:   "abc1",
				MetricType: Concurrency,
				Threshold:  100,
			},
		})
		assert.Nil(t, err)
		for _, r := range GetRules() {
			assert.Equal(t, "123", r.Resource)
		}
		_ = ClearRules()
		WhenUpdateRules(DefaultRuleUpdateHandler)
	})
}

func ruleUpdateForResetResourceHandler(rules []*Rule) (err error) {
	for _, r := range rules {
		r.Resource = "123"
	}
	if err := DefaultRuleUpdateHandler(rules); err != nil {
		return err
	}
	return nil
}
