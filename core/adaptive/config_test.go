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
