package adaptive

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func TestMemoryAdaptiveController(t *testing.T) {
	c1 := &Config{
		AdaptiveConfigName: "test1",
		AdaptiveType:       Memory,
		LowRatio:           1,
		HighRatio:          0.1,
		LowWaterMark:       1024,
		HighWaterMark:      2048,
	}
	mc := newMemoryAdaptiveController(c1)
	system_metric.SetSystemMemoryUsage(100)
	assert.True(t, util.Float64Equals(mc.CalculateSystemAdaptiveCount(1000), 1000))
	system_metric.SetSystemMemoryUsage(1024)
	assert.True(t, util.Float64Equals(mc.CalculateSystemAdaptiveCount(1000), 1000))
	system_metric.SetSystemMemoryUsage(1536)
	assert.True(t, util.Float64Equals(mc.CalculateSystemAdaptiveCount(1000), 550))
	system_metric.SetSystemMemoryUsage(2048)
	assert.True(t, util.Float64Equals(mc.CalculateSystemAdaptiveCount(1000), 100))
	system_metric.SetSystemMemoryUsage(3072)
	assert.True(t, util.Float64Equals(mc.CalculateSystemAdaptiveCount(1000), 100))
	assert.True(t, mc.BoundConfig() == c1)
}
