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

package metrics

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	hostName, _ = os.Hostname()
	psName      = filepath.Base(os.Args[0])
	pid         = os.Getegid()
	pidString   string

	registerMetrics sync.Once
)

func init() {
	if len(hostName) != 0 {
		hostName = "host:" + hostName
	}

	if len(psName) == 0 {
		psName = "ps:" + psName
	}

	if pid != 0 {
		pidString = "pid:" + strconv.Itoa(pid)
	}
}

var (
	CPURatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sentinel_process_cpu_ratio",
			Help: "current process cpu utilization ratio",
		},
		[]string{"host", "process", "cpu", "process_cpu_ratio"},
	)
	ProcessMemorySize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sentinel_process_memory_size",
			Help: "current process memory size in bytes",
		},
		[]string{"host", "process", "pid", "total_memory_size"},
	)
	ResourceFlowThreshold = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sentinel_resource_flow_threshold",
			Help: "resource flow threshold",
		},
		[]string{"host", "resource", "threshold"},
	)

	metrics = []prometheus.Collector{
		CPURatio,
		ProcessMemorySize,
		ResourceFlowThreshold,
	}
)

// SetCPURatio sets the # of current process's cpu ratio
func SetCPURatio(ratio float64) {
	CPURatio.WithLabelValues(hostName, psName, pidString, "process_cpu_ratio").Set(ratio)
}

// SetProcessMemorySize sets the # of current process's memory
func SetProcessMemorySize(size int64) {
	ProcessMemorySize.WithLabelValues(hostName, psName, pidString, "total_memory_size").Set(float64(size))
}

// SetResourceFlowThreshold sets the # of resource threshold
func SetResourceFlowThreshold(resource string, threshold float64) {
	if len(resource) != 0 {
		resource = "rs:" + resource
	}

	ResourceFlowThreshold.WithLabelValues(hostName, resource, "threshold").Set(threshold)
}

func RegisterSentinelMetrics(registry *prometheus.Registry) {
	if registry == nil {
		return
	}
	registerMetrics.Do(func() {
		for _, metric := range metrics {
			registry.MustRegister(metric)
		}
	})
}

type resettable interface {
	Reset()
}

// Reset all metrics to zero
func ResetSentinelMetrics() {
	for _, metric := range metrics {
		rm, ok := metric.(resettable)
		if ok {
			rm.Reset()
		}
	}
}
