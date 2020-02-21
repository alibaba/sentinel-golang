package system

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_recordCpuUsage(t *testing.T) {
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
	assert.Equal(t, notRetrievedValue, CurrentCpuUsage())

	recordCpuUsage(prev, prev)
	assert.Equal(t, 0.0, CurrentCpuUsage())

	recordCpuUsage(prev, cur)
	assert.InEpsilon(t, expected, CurrentCpuUsage(), 0.001)
}
