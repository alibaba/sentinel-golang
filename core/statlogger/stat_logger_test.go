package statlogger

import (
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

type MockStatWriter struct {
	dataChan chan *StatRollingData
}

func (sw *MockStatWriter) WriteAndFlush(srd *StatRollingData) {
	sw.dataChan <- srd
}

func checkLog(t *testing.T, m *MockStatWriter, expectMap map[string]*int64) {
	for i := 0; i < 2; i++ {
		data := <-m.dataChan
		data.mux.Lock()
		for key, value := range data.counter {
			expectValue, ok := expectMap[key]
			if ok {
				assert.True(t, atomic.LoadInt64(value) <= atomic.LoadInt64(expectValue))
				if atomic.LoadInt64(value) == atomic.LoadInt64(expectValue) {
					delete(expectMap, key)
				} else {
					atomic.AddInt64(expectValue, -1*atomic.LoadInt64(value))
				}
			}
		}
		data.mux.Unlock()
		if len(expectMap) <= 0 {
			break
		}
	}
	assert.True(t, len(expectMap) <= 0)
}

func checkRolling(t *testing.T, m *MockStatWriter) {
	data := <-m.dataChan
	now := util.CurrentTimeMillis()
	var ab uint64
	if data.rollingTimeMillis > now {
		ab = data.rollingTimeMillis - now
	} else {
		ab = now - data.rollingTimeMillis
	}
	assert.True(t, ab < 100)
}

func Test_Stat_Logger(t *testing.T) {
	t.Run("Test_stat_logger", func(t *testing.T) {
		loggerName := "test_stat_logger"
		interval := uint64(500)
		testLogger := NewStatLogger(loggerName, 2, interval, 5, 1024)
		testLogger.mux.Lock()
		m := &MockStatWriter{
			dataChan: make(chan *StatRollingData, 100),
		}
		testLogger.writer = m
		testLogger.mux.Unlock()

		testLogger.Stat(2, "test1", "test2")
		expectMap1 := make(map[string]*int64)
		a := new(int64)
		*a = 2
		expectMap1["test1|test2"] = a
		checkLog(t, m, expectMap1)

		testLogger.Stat(1, "test1", "test2")
		testLogger.Stat(1, "test3")
		testLogger.Stat(2, "test3")
		expectMap2 := make(map[string]*int64)
		b := new(int64)
		*b = 1
		expectMap2["test1|test2"] = b
		c := new(int64)
		*c = 3
		expectMap2["test3"] = c
		checkLog(t, m, expectMap2)

		// check interval
		data1 := <-m.dataChan
		data2 := <-m.dataChan
		data3 := <-m.dataChan
		assert.True(t, data2.timeSlot-data1.timeSlot == interval && data3.timeSlot-data2.timeSlot == interval)
		assert.True(t, data1.rollingTimeMillis-data1.timeSlot == interval)
		assert.True(t, data2.rollingTimeMillis-data2.timeSlot == interval)
		assert.True(t, data3.rollingTimeMillis-data3.timeSlot == interval)

		for i := 0; i < 10; i++ {
			checkRolling(t, m)
		}

		// check maxEntryCount
		for i := 0; i < 100; i++ {
			testLogger.Stat(1, "test1"+strconv.Itoa(i))
		}
		sum := 0
		for i := 0; i < 25; i++ {
			data4 := <-m.dataChan
			data4.mux.Lock()
			assert.True(t, len(data4.counter) <= 5)
			sum += len(data4.counter)
			data4.mux.Unlock()
		}
		assert.True(t, sum == 100)
	})

}
