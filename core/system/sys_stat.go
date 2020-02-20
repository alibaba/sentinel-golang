package system

import (
	"github.com/alibaba/sentinel-golang/util"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

const (
	notRetrievedValue float64 = -1
)

var (
	currentLoad     atomic.Value
	currentCpuUsage atomic.Value

	ssStopChan = make(chan struct{})
)

func InitCollector() {
	currentLoad.Store(notRetrievedValue)
	currentCpuUsage.Store(notRetrievedValue)

	ticker := time.NewTicker(1 * time.Second)
	go util.RunWithRecover(func() {
		for {
			select {
			case <-ticker.C:
				retrieveAndUpdateSystemStat()
			case <-ssStopChan:
				ticker.Stop()
				return
			}
		}
	}, logger)
}

func retrieveAndUpdateSystemStat() {
	cpuStats, err := cpu.Times(false)
	if err != nil {
		logger.Warnf("Failed to retrieve current CPU usage: %+v", err)
	}
	loadStat, err := load.Avg()
	if err != nil {
		logger.Warnf("Failed to retrieve current system load: %+v", err)
	}
	if len(cpuStats) > 0 {
		// TODO: calculate the real CPU usage
		// cpuStat := cpuStats[0]
		// currentCpuUsage.Store(cpuStat.User)
	}
	if loadStat != nil {
		currentLoad.Store(loadStat.Load1)
	}
}

func CurrentLoad() float64 {
	r, ok := currentLoad.Load().(float64)
	if !ok {
		return notRetrievedValue
	}
	return r
}

func CurrentCpuUsage() float64 {
	r, ok := currentCpuUsage.Load().(float64)
	if !ok {
		return notRetrievedValue
	}
	return r
}
