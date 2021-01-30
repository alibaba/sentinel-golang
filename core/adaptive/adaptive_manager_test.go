package adaptive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAdaptiveConfigs(t *testing.T) {
	clearData()
	defer clearData()

	t.Run("loadAdaptiveConfigs", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific["sss"] = 1
		specific["123"] = 3

		ok, err := LoadAdaptiveConfigs([]*Config{
			{
				AdaptiveConfigName: "test1",
				AdaptiveType:       Memory,
				LowRatio:           1.7,
				HighRatio:          1.5,
				LowWaterMark:       1000000,
				HighWaterMark:      2000000,
			},
		})
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = LoadAdaptiveConfigs([]*Config{
			{
				AdaptiveConfigName: "test1",
				AdaptiveType:       Memory,
				LowRatio:           1.7,
				HighRatio:          1.5,
				LowWaterMark:       1000000,
				HighWaterMark:      2000000,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = LoadAdaptiveConfigs([]*Config{
			{
				AdaptiveConfigName: "test1",
				AdaptiveType:       Memory,
				LowRatio:           1.8,
				HighRatio:          1.5,
				LowWaterMark:       1000000,
				HighWaterMark:      2000000,
			},
		})
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestIsValidConfig(t *testing.T) {
	badConfig1 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       1,
		LowRatio:           1.8,
		HighRatio:          1.5,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}

	badConfig2 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.3,
		HighRatio:          1.5,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}

	badConfig3 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.8,
		HighRatio:          1.5,
		LowWaterMark:       4000000,
		HighWaterMark:      2000000,
	}
	badConfig4 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.8,
		HighRatio:          1.5,
		LowWaterMark:       0,
		HighWaterMark:      2000000,
	}

	badConfig5 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.8,
		HighRatio:          1.5,
		LowWaterMark:       1,
		HighWaterMark:      0,
	}
	assert.Error(t, IsValidConfig(badConfig1))
	assert.Error(t, IsValidConfig(badConfig2))
	assert.Error(t, IsValidConfig(badConfig3))
	assert.Error(t, IsValidConfig(badConfig4))
	assert.Error(t, IsValidConfig(badConfig5))
}

func TestOnConfigUpdate(t *testing.T) {
	clearData()
	defer clearData()
	config1 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}
	err := onConfigUpdate([]*Config{
		config1,
	})
	assert.True(t, err == nil)
	assert.True(t, len(acMap) == 1)
	assert.True(t, acMap["test1"].BoundConfig().IsEqualsTo(config1))
	err = onConfigUpdate([]*Config{})
	assert.True(t, len(acMap) == 0)
	assert.True(t, err == nil)
}

func clearData() {
	acMap = make(map[string]Controller)
	currentConfigs = make([]*Config, 0)
}
