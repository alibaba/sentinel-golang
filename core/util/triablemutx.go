package util

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const mutexLocked = 1 << iota

type TriableMutex struct {
	sync.Mutex
}

func (tmux *TriableMutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&tmux.Mutex)), 0, mutexLocked)
}
