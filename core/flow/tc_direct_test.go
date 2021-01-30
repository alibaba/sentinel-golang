package flow

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/adaptive"
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func TestDirectTrafficShapingCalculator_CalculateAllowedTokens(t *testing.T) {
	c := &adaptive.Config{
		AdaptiveConfigName: "test",
		AdaptiveType:       adaptive.Memory,
		LowRatio:           1,
		HighRatio:          0.1,
		LowWaterMark:       1024,
		HighWaterMark:      2048,
	}

	ok, err := adaptive.LoadAdaptiveConfigs([]*adaptive.Config{c})
	assert.Nil(t, err)
	assert.True(t, ok)

	r := &Rule{
		Resource:               "abc1",
		Threshold:              1000,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		AdaptiveConfigName:     "test",
	}
	s, err := generateStatFor(r)
	assert.Empty(t, err)
	tsc, _ := NewTrafficShapingController(r, s)
	dc := NewDirectTrafficShapingCalculator(tsc, 1000)
	system_metric.SetSystemMemoryUsage(100)
	assert.True(t, dc.CalculateAllowedTokens(1, 1) == 1000)
	system_metric.SetSystemMemoryUsage(2048)
	assert.True(t, dc.CalculateAllowedTokens(1, 1) == 100)
	system_metric.SetSystemMemoryUsage(1536)
	assert.True(t, util.Float64Equals(dc.CalculateAllowedTokens(1, 1), 550))
	system_metric.SetSystemMemoryUsage(2049)
	assert.True(t, util.Float64Equals(dc.CalculateAllowedTokens(1, 1), 100))
}
