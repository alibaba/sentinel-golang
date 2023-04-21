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
	"encoding/json"
	"fmt"
)

// MetricType indicates the metric type.
type MetricType int32

const (
	MetricTypeUnknown MetricType = iota
	Memory
)

var (
	metricTypeMap = map[MetricType]string{
		MetricTypeUnknown: "MetricTypeUnknown",
		Memory:            "Memory",
	}
	metricTypeExisted = fmt.Errorf("metirc type existed")
)

// RegistryMetricType adds metric type and corresponding description in order.
func RegistryMetricType(metricType MetricType, desc string) error {
	_, exist := metricTypeMap[metricType]
	if exist {
		return metricTypeExisted
	}
	metricTypeMap[metricType] = desc
	return nil
}

func (m MetricType) String() string {
	name, ok := metricTypeMap[m]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", m)
}

type CalculateStrategy int32

const (
	CalculateStrategyUnknown CalculateStrategy = iota
	Linear
)

var (
	calculateStrategyMap = map[CalculateStrategy]string{
		CalculateStrategyUnknown: "CalculateStrategyUnknown",
		Linear:                   "Linear",
	}
	calculateStrategyExisted = fmt.Errorf("calculate strategy existed")
)

// RegistryCalculateStrategy adds metric type and corresponding description in order.
func RegistryCalculateStrategy(calculateStrategy CalculateStrategy, desc string) error {
	_, exist := calculateStrategyMap[calculateStrategy]
	if exist {
		return calculateStrategyExisted
	}
	calculateStrategyMap[calculateStrategy] = desc
	return nil
}

func (c CalculateStrategy) String() string {
	name, ok := calculateStrategyMap[c]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", c)
}

type Config struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// user-defined name to uniquely determine adaptive strategy
	ConfigName               string                    `json:"configName"`
	MetricType               MetricType                `json:"metricType"`
	CalculateStrategy        CalculateStrategy         `json:"calculateStrategy"`
	LinearStrategyParameters *LinearStrategyParameters `json:"LinearStrategyParameters"`
}

type LinearStrategyParameters struct {
	// linear algorithm related parameters
	// limitation: count * LowRatio > count * HighRatio && HighWaterMark > LowWaterMark
	// if the current water mark is less than or equals to LowWaterMark, count == count * LowRatio
	// if the current water mark is more than or equals to HighWaterMark, count == count * HighRatio
	// if  the current memory usage is in (LowWaterMark, HighWaterMark), threshold is in (count * LowRatio, count * HighRatio)
	// when AdaptiveType == Memory, water mark means memory usage and the unit of water mark is bytes.
	LowRatio      float64 `json:"lowRatio"`
	HighRatio     float64 `json:"highRatio"`
	LowWaterMark  float64 `json:"lowWaterMark"`
	HighWaterMark float64 `json:"highWaterMark"`
}

func (c *Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		linearStrategyParameters := "{}"
		if c.CalculateStrategy == Linear && c.LinearStrategyParameters != nil {
			linearStrategyParameters = fmt.Sprintf("{LowRatio=%.2f, HighRatio=%.2f, LowWaterMark=%.2f, HighWaterMark=%.2f}",
				c.LinearStrategyParameters.LowRatio, c.LinearStrategyParameters.HighRatio,
				c.LinearStrategyParameters.LowWaterMark, c.LinearStrategyParameters.HighWaterMark)
		}
		// Return the fallback string
		return fmt.Sprintf("Config{ConfigName=%s, MetricType=%s, CalculateStrategy=%s, LinearStrategyParameters=%s}",
			c.ConfigName, c.MetricType, c.CalculateStrategy, linearStrategyParameters)
	}
	return string(b)
}
