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

	"github.com/alibaba/sentinel-golang/util"
)

// AdaptiveType indicates the adaptive type.
type AdaptiveType int32

const (
	Memory AdaptiveType = iota
)

func (s AdaptiveType) String() string {
	switch s {
	case Memory:
		return "Memory"
	default:
		return "Undefined"
	}
}

type Config struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// user-defined name to uniquely determine adaptive strategy
	AdaptiveConfigName string       `json:"adaptiveConfigName"`
	AdaptiveType       AdaptiveType `json:"adaptiveType"`

	// adaptive algorithm related parameters
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

func (c *Config) IsEqualsTo(newConfig *Config) bool {
	if newConfig == nil {
		return false
	}
	return c.AdaptiveConfigName == newConfig.AdaptiveConfigName && c.AdaptiveType == newConfig.AdaptiveType &&
		util.Float64Equals(c.LowRatio, newConfig.LowRatio) && util.Float64Equals(c.HighRatio, newConfig.HighRatio) &&
		util.Float64Equals(c.LowWaterMark, newConfig.LowWaterMark) && util.Float64Equals(c.HighWaterMark, newConfig.HighWaterMark)
}

func (c *Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Config{AdaptiveConfigName=%s, AdaptiveType=%s, LowRatio=%.2f, "+
			"HighRatio=%.2f, LowWaterMark=%.2f, HighWaterMark=%.2f}",
			c.AdaptiveConfigName, c.AdaptiveType, c.LowRatio, c.HighRatio, c.LowWaterMark, c.HighWaterMark)
	}
	return string(b)
}
