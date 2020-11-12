package statlogger

import (
	"strconv"
	"sync"
	"time"

	"github.com/alibaba/sentinel-golang/logging"

	"github.com/alibaba/sentinel-golang/util"
)

var statLoggers = make(map[string]*StatLogger)

var mux = new(sync.Mutex)

const (
	logFlushQueueSize = 60
)

// NewStatLogger constructs a NewStatLogger
func NewStatLogger(loggerName string, maxBackupIndex int, intervalMillis uint64, maxEntryCount int, maxFileSize uint64) *StatLogger {
	sw, err := NewStatFileWriter(loggerName, maxFileSize, maxBackupIndex)
	if err != nil {
		return nil
	}
	sl := &StatLogger{
		loggerName:     loggerName,
		intervalMillis: intervalMillis,
		maxEntryCount:  maxEntryCount,
		mux:            new(sync.Mutex),
		writeChan:      make(chan *StatRollingData, logFlushQueueSize),
		rollingChan:    make(chan int),
		writer:         sw,
	}
	sl.Rolling()
	// Schedule the log flushing task
	go util.RunWithRecover(sl.WriteTaskLoop)
	addLogger(sl)
	return sl
}

func addLogger(sl *StatLogger) *StatLogger {
	mux.Lock()
	defer mux.Unlock()
	logger, ok := statLoggers[sl.loggerName]
	if ok {
		return logger
	}

	statLoggers[sl.loggerName] = sl
	go util.RunWithRecover(func() {
		for {
			select {
			case <-sl.rollingChan:
				sl.writeChan <- sl.Rolling()
				go nextRolling(sl)
			}
		}
	})
	go nextRolling(sl)
	return sl
}

func nextRolling(sl *StatLogger) {
	rollingTimeMillis := sl.data.Load().(*StatRollingData).rollingTimeMillis
	delayMillis := int64(rollingTimeMillis) - int64(util.CurrentTimeMillis())
	if delayMillis > 5 {
		timer := time.NewTimer(time.Duration(delayMillis) * time.Millisecond)
		<-timer.C
		sl.rollingChan <- 1
	} else if -delayMillis > int64(sl.intervalMillis) {
		logging.Warn("[StatLogController] unusual delay of statLogger[" + sl.loggerName + "], " +
			"delay=" + strconv.FormatInt(-delayMillis, 10) + "ms, submit now")
		sl.rollingChan <- 1
	} else {
		sl.rollingChan <- 1
	}
}
