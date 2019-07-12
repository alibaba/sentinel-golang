package util

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTriableMutex_TryLock(t *testing.T) {
	var m TriableMutex
	m.Lock()
	time.Sleep(time.Second)
	fmt.Printf("TryLock: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock: %t\n", m.TryLock()) // false
	m.Unlock()
	fmt.Printf("TryLock: %t\n", m.TryLock()) //true
	fmt.Printf("TryLock: %t\n", m.TryLock()) //false
	m.Unlock()
	fmt.Printf("TryLock: %t\n", m.TryLock()) //true
	m.Unlock()
}

func iiiTestConcurrent100() {
	var m TriableMutex
	cnt := int32(0)
	wg := &sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		ci := i
		go func(tm TriableMutex, wg_ *sync.WaitGroup, i_ int, cntPtr *int32) {
			for {
				if tm.TryLock() {
					atomic.AddInt32(cntPtr, 1)
					tm.Unlock()
					wg_.Done()
					break
				} else {
					runtime.Gosched()
				}
			}
		}(m, wg, ci, &cnt)
	}
	wg.Wait()
	//fmt.Println("count=", cnt)
	if cnt != 1000 {
		fmt.Println("count error")
	}
}

func TestTriableMutex_TryLock_Concurrent1000(t *testing.T) {
	iiiTestConcurrent100()
}

func BenchmarkTriableMutex_TryLock(b *testing.B) {
	for n := 0; n < b.N; n++ {
		iiiTestConcurrent100()
	}
}
