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

package system_metric

import (
	"testing"

	"github.com/alibaba/sentinel-golang/util"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
)

func Test_recordCpuUsage(t *testing.T) {
	defer currentCpuUsage.Store(NotRetrievedValue)

	var emptyStat *cpu.TimesStat = nil
	// total: 2260, user+nice: 950, system+irqs=210
	prev := &cpu.TimesStat{
		CPU:     "all",
		User:    900,
		System:  200,
		Idle:    300,
		Nice:    50,
		Iowait:  100,
		Irq:     5,
		Softirq: 5,
		Steal:   700,
	}
	// total: 4180, user+nice: 1600, system+irqs=430
	cur := &cpu.TimesStat{
		CPU:     "all",
		User:    1500,
		System:  400,
		Idle:    400,
		Nice:    100,
		Iowait:  150,
		Irq:     15,
		Softirq: 15,
		Steal:   1600,
	}
	expected := float64(1600+430-950-210) / (4180 - 2260)

	recordCpuUsage(emptyStat, cur)
	assert.True(t, util.Float64Equals(NotRetrievedValue, CurrentCpuUsage()))

	recordCpuUsage(prev, prev)
	assert.True(t, util.Float64Equals(0.0, CurrentCpuUsage()))

	recordCpuUsage(prev, cur)
	assert.InEpsilon(t, expected, CurrentCpuUsage(), 0.001)
}

func TestCurrentLoad(t *testing.T) {
	defer currentLoad.Store(NotRetrievedValue)

	cLoad := CurrentLoad()
	assert.True(t, util.Float64Equals(NotRetrievedValue, cLoad))

	v := float64(1.0)
	currentLoad.Store(v)
	cLoad = CurrentLoad()
	assert.True(t, util.Float64Equals(v, cLoad))
}

func TestCurrentCpuUsage(t *testing.T) {
	defer currentCpuUsage.Store(NotRetrievedValue)

	cpuUsage := CurrentCpuUsage()
	assert.Equal(t, NotRetrievedValue, cpuUsage)

	v := float64(0.3)
	currentCpuUsage.Store(v)
	cpuUsage = CurrentCpuUsage()
	assert.True(t, util.Float64Equals(v, cpuUsage))
}
