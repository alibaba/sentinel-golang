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

package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricTypeString(t *testing.T) {
	t.Run("LoadMetricType", func(t *testing.T) {
		mt := Load
		assert.Equal(t, "load", mt.String())
	})

	t.Run("AvgRTMetricType", func(t *testing.T) {
		mt := AvgRT
		assert.Equal(t, "avgRT", mt.String())
	})

	t.Run("ConcurrencyMetricType", func(t *testing.T) {
		mt := Concurrency
		assert.Equal(t, "concurrency", mt.String())
	})

	t.Run("InboundQPSMetricType", func(t *testing.T) {
		mt := InboundQPS
		assert.Equal(t, "inboundQPS", mt.String())
	})

	t.Run("CpuUsageQPSMetricType", func(t *testing.T) {
		mt := CpuUsage
		assert.Equal(t, "cpuUsage", mt.String())
	})

	t.Run("UnknownMetricType", func(t *testing.T) {
		mt := MetricTypeSize
		assert.Equal(t, "unknown(5)", mt.String())
	})
}

func TestAdaptiveStrategyString(t *testing.T) {
	t.Run("NoAdaptiveStrategy", func(t *testing.T) {
		as := NoAdaptive
		assert.Equal(t, "none", as.String())
	})

	t.Run("BBRAdaptiveStrategy", func(t *testing.T) {
		as := BBR
		assert.Equal(t, "bbr", as.String())
	})

	t.Run("UnknownAdaptiveStrategy", func(t *testing.T) {
		as := AdaptiveStrategy(2)
		assert.Equal(t, "unknown(2)", as.String())
	})
}

func TestSystemRuleResourceName(t *testing.T) {
	t.Run("ValidResourceName", func(t *testing.T) {
		sr := &Rule{MetricType: Concurrency}
		assert.Equal(t, "concurrency", sr.ResourceName())
	})
}

func TestSystemRuleString(t *testing.T) {
	t.Run("ValidSystemRuleString", func(t *testing.T) {
		sr := &Rule{MetricType: Concurrency}
		assert.NotContains(t, sr.String(), "Rule")
	})
}
