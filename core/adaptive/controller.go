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
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/logging"
)

type Controller interface {
	CalculateSystemAdaptiveCount(count float64) float64

	BoundConfig() *Config
}

type BaseController struct {
	config *Config
}

func newBaseController(c *Config) *BaseController {
	return &BaseController{
		config: c,
	}
}

func (bc BaseController) BoundConfig() *Config {
	return bc.config
}

// MemoryAdaptiveController is a memory adaptive controller.
//
// adaptive flow control algorithm
// If the watermark is less than Config.memLowWaterMark, the count is Config.lowMemUsageRatio * count.
// If the watermark is greater than Config.memHighWaterMark, the count is Config.highMemUsageRatio * count.
// Otherwise, the count is ((watermark - memLowWaterMark)/(memHighWaterMark - memLowWaterMark)) *
//	(highMemUsageRatio * count - lowMemUsageRatio * count) + lowMemUsageRatio * count.
type MemoryAdaptiveController struct {
	BaseController
	lowMemUsageRatio  float64
	highMemUsageRatio float64
	memLowWaterMark   int64
	memHighWaterMark  int64
}

func newMemoryAdaptiveController(c *Config) *MemoryAdaptiveController {
	return &MemoryAdaptiveController{
		BaseController:    *newBaseController(c),
		lowMemUsageRatio:  c.LowRatio,
		highMemUsageRatio: c.HighRatio,
		memLowWaterMark:   int64(c.LowWaterMark),
		memHighWaterMark:  int64(c.HighWaterMark),
	}
}

func (mc *MemoryAdaptiveController) CalculateSystemAdaptiveCount(count float64) float64 {
	var adaptiveCount float64
	lowMemUsageCount := count * mc.lowMemUsageRatio
	highMemUsageCount := count * mc.highMemUsageRatio
	mem := system_metric.CurrentMemoryUsage()
	if mem == system_metric.NotRetrievedMemoryValue {
		logging.Warn("[MemoryAdaptiveController CalculateSystemAdaptiveCount]Fail to load memory usage")
		return lowMemUsageCount
	}
	if mem <= mc.memLowWaterMark {
		adaptiveCount = lowMemUsageCount
	} else if mem >= mc.memHighWaterMark {
		adaptiveCount = highMemUsageCount
	} else {
		adaptiveCount = ((highMemUsageCount-lowMemUsageCount)/float64(mc.memHighWaterMark-mc.memLowWaterMark))*float64(mem-mc.memLowWaterMark) + lowMemUsageCount
	}
	return adaptiveCount
}
