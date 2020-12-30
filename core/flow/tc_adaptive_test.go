package flow

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func TestMemoryAdaptiveTrafficShapingCalculator_CalculateAllowedTokens(t *testing.T) {
	tc1 := &MemoryAdaptiveTrafficShapingCalculator{
		owner:                 nil,
		lowMemUsageThreshold:  1000,
		highMemUsageThreshold: 100,
		memLowWaterMark:       1024,
		memHighWaterMark:      2048,
	}
	system_metric.SetSystemMemoryUsage(100)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), float64(tc1.lowMemUsageThreshold)))
	system_metric.SetSystemMemoryUsage(1024)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), float64(tc1.lowMemUsageThreshold)))
	system_metric.SetSystemMemoryUsage(1536)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 550))
	system_metric.SetSystemMemoryUsage(2048)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 100))
	system_metric.SetSystemMemoryUsage(3072)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 100))
}
