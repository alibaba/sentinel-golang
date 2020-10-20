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
		counter:           make(map[string]uint32, initCap),
		mux:               new(sync.Mutex),
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
	counter           map[string]uint32
	mux               *sync.Mutex
	sl                *StatLogger
}

func (s *StatRollingData) CountAndSum(args []string, count uint32) {
	s.mux.Lock()
	defer s.mux.Unlock()
	key := strings.Join(args, "|")
	num, ok := s.counter[key]
	if !ok {
		num = 0
		size := len(s.counter)
		// When entry size bigger than maxEntryCount
		if size < s.sl.maxEntryCount {
			s.counter[key] = num
		} else {
			old := s.counter
			s.counter = make(map[string]uint32, 16)
			clone := StatRollingData{
				timeSlot:          s.timeSlot,
				rollingTimeMillis: s.rollingTimeMillis,
				counter:           old,
				mux:               new(sync.Mutex),
				sl:                s.sl,
			}
			s.sl.writeChan <- &clone
		}
	}
	s.counter[key] = num + count
}

func (s *StatRollingData) GetCloneDataAndClear() map[string]uint32 {
	s.mux.Lock()
	defer s.mux.Unlock()
	counter := s.counter
	s.counter = make(map[string]uint32)
	return counter
}

func (s *StatRollingData) Len() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return len(s.counter)
}
