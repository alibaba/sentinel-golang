package system

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsValidSystemRule(t *testing.T) {
	badRule1 := &SystemRule{MetricType: InboundQPS, TriggerCount: -1}
	badRule2 := &SystemRule{MetricType: CpuUsage, TriggerCount: 75}
	goodRule1 := &SystemRule{MetricType: Load, TriggerCount: 12, Strategy: BBR}

	assert.Error(t, IsValidSystemRule(badRule1))
	assert.Error(t, IsValidSystemRule(badRule2))
	assert.NoError(t, IsValidSystemRule(goodRule1))
}
