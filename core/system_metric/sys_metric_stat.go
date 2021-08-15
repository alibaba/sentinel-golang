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
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/metrics"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/atomic"
)

const (
	NotRetrievedLoadValue     float64 = -1.0
	NotRetrievedCpuUsageValue float64 = -1.0
	NotRetrievedMemoryValue   int64   = -1
	CGroupPath                        = "/proc/1/cgroup"
	DockerPath                        = "/docker"
	KubepodsPath                      = "/kubepods"
)

var (
	currentLoad        atomic.Value
	currentCpuUsage    atomic.Value
	currentMemoryUsage atomic.Value

	loadStatCollectorOnce   sync.Once
	memoryStatCollectorOnce sync.Once
	cpuStatCollectorOnce    sync.Once

	CurrentPID         = os.Getpid()
	currentProcess     atomic.Value
	currentProcessOnce sync.Once
	TotalMemorySize    = getTotalMemorySize()

	ssStopChan = make(chan struct{})

	isContainer             bool
	preSysTotalCpuUsage     atomic.Float64
	preContainerCpuUsage    atomic.Float64
	onlineContainerCpuCount float64
)

func init() {
	currentLoad.Store(NotRetrievedLoadValue)
	currentCpuUsage.Store(NotRetrievedCpuUsageValue)
	currentMemoryUsage.Store(NotRetrievedMemoryValue)

	p, err := process.NewProcess(int32(CurrentPID))
	if err != nil {
		logging.Error(err, "Fail to new process when initializing system metric", "pid", CurrentPID)
		return
	}
	currentProcessOnce.Do(func() {
		currentProcess.Store(p)
	})

	isContainer = isContainerRunning()
	if isContainer {
		var (
			currentSysCpuTotal       float64
			currentContainerCpuTotal float64
		)

		currentSysCpuTotal, err = getSysCpuUsage()
		if err != nil {
			logging.Error(err, "Fail to getSysCpuUsage when initializing system metric")
			return
		}
		currentContainerCpuTotal, err = getContainerCpuUsage()
		if err != nil {
			logging.Error(err, "Fail to getContainerCpuUsage when initializing system metric")
			return
		}
		preContainerCpuUsage.Store(currentContainerCpuTotal)
		preSysTotalCpuUsage.Store(currentSysCpuTotal)
		onlineContainerCpuCount = getContainerCpuCount()
	}
}

func readLineFromFile(filepath string) []string {
	res := make([]string, 0)
	f, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer f.Close()
	buff := bufio.NewReader(f)
	for {
		line, _, err := buff.ReadLine()
		if err != nil {
			return res
		}
		res = append(res, string(line))
	}
}

func isContainerRunning() bool {

	lines := readLineFromFile(CGroupPath)
	for _, line := range lines {
		if strings.Contains(line, DockerPath) ||
			strings.Contains(line, KubepodsPath) {
			return true
		}
	}
	return false
}

func getContainerCpuCount() float64 {
	path := "/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu"
	usage := readLineFromFile(path)
	if len(usage) == 0 {
		return 0
	}
	perCpuUsages := strings.Fields(usage[0])

	return float64(len(perCpuUsages))
}

// getMemoryStat returns the current machine's memory statistic
func getTotalMemorySize() (total uint64) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		logging.Error(err, "Fail to read Virtual Memory")
		return 0
	}
	return stat.Total
}

func InitMemoryCollector(intervalMs uint32) {
	if intervalMs == 0 {
		return
	}
	memoryStatCollectorOnce.Do(func() {
		// Initial memory retrieval.
		retrieveAndUpdateMemoryStat()

		ticker := util.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		go util.RunWithRecover(func() {
			for {
				select {
				case <-ticker.C():
					retrieveAndUpdateMemoryStat()
				case <-ssStopChan:
					ticker.Stop()
					return
				}
			}
		})
	})
}

func retrieveAndUpdateMemoryStat() {
	var (
		memoryUsedBytes int64
		err             error
	)
	if isContainer {
		memoryUsedBytes = GetContainerMemoryStat()
	} else {
		memoryUsedBytes, err = GetProcessMemoryStat()
		if err != nil {
			logging.Error(err, "Fail to retrieve and update memory statistic")
			return
		}
	}
	metrics.SetProcessMemorySize(memoryUsedBytes)
	currentMemoryUsage.Store(memoryUsedBytes)
}

func GetContainerMemoryStat() int64 {
	path := "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	usage := readLineFromFile(path)
	if len(usage) == 0 {
		return 0
	}
	ns, err := strconv.ParseInt(strings.TrimSpace(usage[0]), 10, 64)
	if err != nil {
		return 0
	}
	return ns
}

// GetProcessMemoryStat gets current process's memory usage in Bytes
func GetProcessMemoryStat() (int64, error) {
	curProcess := currentProcess.Load()
	if curProcess == nil {
		p, err := process.NewProcess(int32(CurrentPID))
		if err != nil {
			return 0, err
		}
		currentProcessOnce.Do(func() {
			currentProcess.Store(p)
		})
		curProcess = currentProcess.Load()
	}
	p := curProcess.(*process.Process)
	memInfo, err := p.MemoryInfo()
	var rss int64
	if memInfo != nil {
		rss = int64(memInfo.RSS)
	}

	return rss, err
}

func InitCpuCollector(intervalMs uint32) {
	if intervalMs == 0 {
		return
	}
	cpuStatCollectorOnce.Do(func() {
		// Initial memory retrieval.
		retrieveAndUpdateCpuStat()

		ticker := util.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		go util.RunWithRecover(func() {
			for {
				select {
				case <-ticker.C():
					retrieveAndUpdateCpuStat()
				case <-ssStopChan:
					ticker.Stop()
					return
				}
			}
		})
	})
}

func retrieveAndUpdateCpuStat() {
	var (
		cpuPercent float64
		err        error
	)
	if isContainer {
		cpuPercent, err = GetContainerCpuStat()
		if err != nil {
			logging.Error(err, "Fail to retrieve and update cpu statistic")
			return
		}
	} else {
		cpuPercent, err = getProcessCpuStat()
		if err != nil {
			logging.Error(err, "Fail to retrieve and update cpu statistic")
			return
		}
	}

	metrics.SetCPURatio(cpuPercent)
	currentCpuUsage.Store(cpuPercent)
}

func GetContainerCpuStat() (float64, error) {

	var (
		currentSysCpuTotal       float64
		currentContainerCpuTotal float64
		err                      error
	)

	currentSysCpuTotal, err = getSysCpuUsage()
	if err != nil {
		return 0, err
	}
	currentContainerCpuTotal, err = getContainerCpuUsage()
	if err != nil {
		return 0, err
	}

	preSysTotalCpu := preSysTotalCpuUsage.Load()

	preContainerCpu := preContainerCpuUsage.Load()

	preSysTotalCpuUsage.Store(currentSysCpuTotal)
	preContainerCpuUsage.Store(currentContainerCpuTotal)

	if currentSysCpuTotal-preSysTotalCpu == 0 {
		return 0, err
	}
	return (currentContainerCpuTotal - preContainerCpu) * onlineContainerCpuCount / (currentSysCpuTotal - preSysTotalCpu), err
}

func getSysCpuUsage() (float64, error) {
	var (
		currentSysCpuTotal float64
	)
	currentCpuStatArr, err := cpu.Times(false)
	if err != nil {
		return 0, err
	}
	for _, stat := range currentCpuStatArr {
		currentSysCpuTotal = stat.User + stat.System + stat.Idle + stat.Nice + stat.Iowait + stat.Irq +
			stat.Softirq + stat.Steal + stat.Guest + stat.GuestNice
	}
	return currentSysCpuTotal, nil
}

func getContainerCpuUsage() (float64, error) {
	path := "/sys/fs/cgroup/cpuacct/cpuacct.usage"
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	usage, _, err := reader.ReadLine()
	if err != nil {
		return 0, err
	}
	ns, err := strconv.ParseFloat(strings.TrimSpace(string(usage)), 64)
	if err != nil {
		return 0, err
	}
	return ns / 1e9, nil
}

// getProcessCpuStat gets current process's cpu usage in Bytes
func getProcessCpuStat() (float64, error) {
	curProcess := currentProcess.Load()
	if curProcess == nil {
		p, err := process.NewProcess(int32(CurrentPID))
		if err != nil {
			return 0, err
		}
		currentProcessOnce.Do(func() {
			currentProcess.Store(p)
		})
		curProcess = currentProcess.Load()
	}
	p := curProcess.(*process.Process)
	return p.Percent(0)
}

func InitLoadCollector(intervalMs uint32) {
	if intervalMs == 0 {
		return
	}
	loadStatCollectorOnce.Do(func() {
		// Initial retrieval.
		retrieveAndUpdateLoadStat()

		ticker := util.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		go util.RunWithRecover(func() {
			for {
				select {
				case <-ticker.C():
					retrieveAndUpdateLoadStat()
				case <-ssStopChan:
					ticker.Stop()
					return
				}
			}
		})
	})
}

func retrieveAndUpdateLoadStat() {
	loadStat, err := load.Avg()
	if err != nil {
		logging.Error(err, "[retrieveAndUpdateSystemStat] Failed to retrieve current system load")
		return
	}
	if loadStat != nil {
		currentLoad.Store(loadStat.Load1)
	}
}

func CurrentLoad() float64 {
	r, ok := currentLoad.Load().(float64)
	if !ok {
		return NotRetrievedLoadValue
	}
	return r
}

// Note: SetSystemLoad is used for unit test, the user shouldn't call this function.
func SetSystemLoad(load float64) {
	currentLoad.Store(load)
}

func CurrentCpuUsage() float64 {
	r, ok := currentCpuUsage.Load().(float64)
	if !ok {
		return NotRetrievedCpuUsageValue
	}
	return r
}

// Note: SetSystemCpuUsage is used for unit test, the user shouldn't call this function.
func SetSystemCpuUsage(cpuUsage float64) {
	currentCpuUsage.Store(cpuUsage)
}

func CurrentMemoryUsage() int64 {
	bytes, ok := currentMemoryUsage.Load().(int64)
	if !ok {
		return NotRetrievedMemoryValue
	}

	return bytes
}

// Note: SetSystemCpuUsage is used for unit test, the user shouldn't call this function.
func SetSystemMemoryUsage(memoryUsage int64) {
	currentMemoryUsage.Store(memoryUsage)
}
