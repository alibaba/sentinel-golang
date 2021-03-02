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
				ConfigName:        "test1",
				MetricType:        Memory,
				CalculateStrategy: Linear,
				LinearStrategyParameters: &LinearStrategyParameters{
					LowRatio:      1.7,
					HighRatio:     1.5,
					LowWaterMark:  1000000,
					HighWaterMark: 2000000,
				},
			},
		})
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = LoadAdaptiveConfigs([]*Config{
			{
				ConfigName:        "test1",
				MetricType:        Memory,
				CalculateStrategy: Linear,
				LinearStrategyParameters: &LinearStrategyParameters{
					LowRatio:      1.7,
					HighRatio:     1.5,
					LowWaterMark:  1000000,
					HighWaterMark: 2000000,
				},
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = LoadAdaptiveConfigs([]*Config{
			{
				ConfigName:        "test1",
				MetricType:        Memory,
				CalculateStrategy: Linear,
				LinearStrategyParameters: &LinearStrategyParameters{
					LowRatio:      1.8,
					HighRatio:     1.5,
					LowWaterMark:  1000000,
					HighWaterMark: 2000000,
				},
			},
		})
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestIsValidConfig(t *testing.T) {
	badConfig1 := &Config{
		ConfigName:        "test1",
		MetricType:        2,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.8,
			HighRatio:     1.5,
			LowWaterMark:  1000000,
			HighWaterMark: 2000000,
		},
	}

	badConfig2 := &Config{
		ConfigName:        "test1",
		MetricType:        Memory,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.3,
			HighRatio:     1.5,
			LowWaterMark:  1000000,
			HighWaterMark: 2000000,
		},
	}

	badConfig3 := &Config{
		ConfigName:        "test1",
		MetricType:        Memory,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.8,
			HighRatio:     1.5,
			LowWaterMark:  4000000,
			HighWaterMark: 2000000,
		},
	}
	badConfig4 := &Config{
		ConfigName:        "test1",
		MetricType:        Memory,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.8,
			HighRatio:     1.5,
			LowWaterMark:  0,
			HighWaterMark: 2000000,
		},
	}

	badConfig5 := &Config{
		ConfigName:        "test1",
		MetricType:        Memory,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.8,
			HighRatio:     1.5,
			LowWaterMark:  1,
			HighWaterMark: 0,
		},
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
		ConfigName:        "test1",
		MetricType:        Memory,
		CalculateStrategy: Linear,
		LinearStrategyParameters: &LinearStrategyParameters{
			LowRatio:      1.7,
			HighRatio:     1.5,
			LowWaterMark:  1000000,
			HighWaterMark: 2000000,
		},
	}
	err := onConfigUpdate([]*Config{
		config1,
	})
	assert.True(t, err == nil)
	assert.True(t, len(acMap) == 1)
	err = onConfigUpdate([]*Config{})
	assert.True(t, len(acMap) == 0)
	assert.True(t, err == nil)
}

func clearData() {
	acMap = make(map[string]Controller)
	currentConfigs = make([]*Config, 0)
}
