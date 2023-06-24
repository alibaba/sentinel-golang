package flow

import (
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/logging"
)

// MemoryAdaptiveTrafficShapingCalculator is a memory adaptive traffic shaping calculator
//
// adaptive flow control algorithm
// If the watermark is less than Rule.MemLowWaterMarkBytes, the threshold is Rule.LowMemUsageThreshold.
// If the watermark is greater than Rule.MemHighWaterMarkBytes, the threshold is Rule.HighMemUsageThreshold.
// Otherwise, the threshold is ((watermark - MemLowWaterMarkBytes)/(MemHighWaterMarkBytes - MemLowWaterMarkBytes)) *
//
//	(HighMemUsageThreshold - LowMemUsageThreshold) + LowMemUsageThreshold.
type MemoryAdaptiveTrafficShapingCalculator struct {
	owner                 *TrafficShapingController
	lowMemUsageThreshold  int64
	highMemUsageThreshold int64
	memLowWaterMark       int64
	memHighWaterMark      int64
}

func NewMemoryAdaptiveTrafficShapingCalculator(owner *TrafficShapingController, r *Rule) *MemoryAdaptiveTrafficShapingCalculator {
	return &MemoryAdaptiveTrafficShapingCalculator{
		owner:                 owner,
		lowMemUsageThreshold:  r.LowMemUsageThreshold,
		highMemUsageThreshold: r.HighMemUsageThreshold,
		memLowWaterMark:       r.MemLowWaterMarkBytes,
		memHighWaterMark:      r.MemHighWaterMarkBytes,
	}
}

func (m *MemoryAdaptiveTrafficShapingCalculator) BoundOwner() *TrafficShapingController {
	return m.owner
}

func (m *MemoryAdaptiveTrafficShapingCalculator) CalculateAllowedTokens(_ uint32, _ int32) float64 {
	var threshold float64
	mem := system_metric.CurrentMemoryUsage()
	if mem == system_metric.NotRetrievedMemoryValue {
		logging.Warn("[MemoryAdaptiveTrafficShapingCalculator CalculateAllowedTokens]Fail to load memory usage")
		return float64(m.lowMemUsageThreshold)
	}
	if mem <= m.memLowWaterMark {
		threshold = float64(m.lowMemUsageThreshold)
	} else if mem >= m.memHighWaterMark {
		threshold = float64(m.highMemUsageThreshold)
	} else {
		threshold = (float64(m.highMemUsageThreshold-m.lowMemUsageThreshold)/float64(m.memHighWaterMark-m.memLowWaterMark))*float64(mem-m.memLowWaterMark) + float64(m.lowMemUsageThreshold)
	}
	return threshold
}
