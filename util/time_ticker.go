package util

import (
	"sync/atomic"
	"time"
)

var nowInMs = uint64(0)

func StartTimeTicker() {
	atomic.StoreUint64(&nowInMs, uint64(time.Now().UnixNano())/UnixTimeUnitOffset)
	go func() {
		for {
			now := uint64(time.Now().UnixNano()) / UnixTimeUnitOffset
			atomic.StoreUint64(&nowInMs, now)
			time.Sleep(time.Millisecond)
		}
	}()
}

func CurrentTimeMillWithTicker() uint64 {
	return atomic.LoadUint64(&nowInMs)
}
