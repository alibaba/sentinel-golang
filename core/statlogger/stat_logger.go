package statlogger

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/util"
)

type StatLogger struct {
	loggerName     string
	intervalMillis uint64
	maxEntryCount  int
	data           atomic.Value
	writeChan      chan *StatRollingData
	writer         StatWriter
	mux            *sync.Mutex
	rollingChan    chan int
}

// Stat stats the args within the period.
func (s *StatLogger) Stat(count uint32, args ...string) {
	s.data.Load().(*StatRollingData).CountAndSum(args, count)
}

// WriteTaskLoop begins writing loop.
func (s *StatLogger) WriteTaskLoop() {
	for {
		select {
		case srd := <-s.writeChan:
			s.writer.WriteAndFlush(srd)
		}
	}
}

// Rolling Rolls StatLogger to next statistical period.
func (s *StatLogger) Rolling() *StatRollingData {
	s.mux.Lock()
	defer s.mux.Unlock()

	prevData := s.data.Load()
	var timeSlot, rollingTimeMillis uint64
	var initCap int

	if prevData == nil {
		now := util.CurrentTimeMillis()
		timeSlot = now - now%s.intervalMillis
		rollingTimeMillis = timeSlot + s.intervalMillis
		initCap = 16
	} else {
		prevRollingData := prevData.(*StatRollingData)
		now := util.CurrentTimeMillis()
		timeSlot = now - now%s.intervalMillis
		if timeSlot <= prevRollingData.timeSlot {
			timeSlot = prevRollingData.timeSlot + s.intervalMillis
		}
		rollingTimeMillis = timeSlot + s.intervalMillis
		initCap = prevRollingData.Len()
	}

	sr := StatRollingData{
		timeSlot:          timeSlot,
		rollingTimeMillis: rollingTimeMillis,
		counter:           make(map[string]*int64, initCap),
		mux:               new(sync.RWMutex),
		sl:                s,
	}
	s.data.Store(&sr)
	if prevData == nil {
		return nil
	}
	return prevData.(*StatRollingData)

}

type StatRollingData struct {
	timeSlot          uint64
	rollingTimeMillis uint64
	counter           map[string]*int64
	mux               *sync.RWMutex
	sl                *StatLogger
}

func (s *StatRollingData) CountAndSum(args []string, count uint32) {
	key := strings.Join(args, "|")
	s.mux.RLock()
	if s.counter == nil {
		s.mux.RUnlock()
		s.sl.Stat(count, args...)
		return
	}
	num, ok := s.counter[key]
	if ok {
		atomic.AddInt64(num, int64(count))
		s.mux.RUnlock()
		return
	}
	s.mux.RUnlock()

	s.mux.Lock()
	if s.counter == nil {
		s.mux.Unlock()
		s.sl.Stat(count, args...)
		return
	}
	num, ok = s.counter[key]
	if ok {
		atomic.AddInt64(num, int64(count))
		s.mux.Unlock()
		return
	}
	num = new(int64)
	*num = int64(count)
	size := len(s.counter)
	// When entry size bigger than maxEntryCount
	if size < s.sl.maxEntryCount {
		s.counter[key] = num
		s.mux.Unlock()
	} else {
		old := s.counter
		s.counter = make(map[string]*int64, 8)
		clone := StatRollingData{
			timeSlot:          s.timeSlot,
			rollingTimeMillis: s.rollingTimeMillis,
			counter:           old,
			mux:               new(sync.RWMutex),
			sl:                s.sl,
		}
		s.counter[key] = num
		s.mux.Unlock()
		s.sl.writeChan <- &clone
	}
}

func (s *StatRollingData) GetCloneDataAndClear() map[string]*int64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	counter := s.counter
	s.counter = nil
	return counter
}

func (s *StatRollingData) Len() int {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.counter)
}
