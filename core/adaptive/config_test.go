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

func TestConfigIsEqualsTo(t *testing.T) {
	c1 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}
	c2 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}
	c3 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.51,
		LowWaterMark:       1000000,
		HighWaterMark:      2000000,
	}
	c4 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       500000,
		HighWaterMark:      2000000,
	}
	c5 := &Config{
		AdaptiveConfigName: "test2",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       500000,
		HighWaterMark:      2000000,
	}
	c6 := &Config{
		AdaptiveConfigName: "test2",
		AdaptiveType:       Memory,
		LowRatio:           1.7,
		HighRatio:          1.5,
		LowWaterMark:       500000,
		HighWaterMark:      7000000,
	}
	c7 := &Config{
		AdaptiveConfigName: "test2",
		AdaptiveType:       Memory,
		LowRatio:           1.8,
		HighRatio:          1.5,
		LowWaterMark:       500000,
		HighWaterMark:      7000000,
	}

	assert.True(t, c1.IsEqualsTo(c2))
	assert.False(t, c1.IsEqualsTo(c3))
	assert.False(t, c1.IsEqualsTo(c4))
	assert.False(t, c1.IsEqualsTo(c5))
	assert.False(t, c1.IsEqualsTo(c6))
	assert.False(t, c1.IsEqualsTo(c7))
}
