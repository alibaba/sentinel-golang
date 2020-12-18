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

package base

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/util"
)

func Test_Mutex_TryLock(t *testing.T) {
	var m mutex
	m.Lock()
	util.Sleep(time.Second)
	if m.TryLock() {
		t.Error("TryLock get lock error")
	}
	m.Unlock()
	if !m.TryLock() {
		t.Error("TryLock get lock error")
	}
	m.Unlock()
}

func utTriableMutexConcurrent(t *testing.T) {
	m := &mutex{}
	cnt := int32(0)
	wg := &sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(tm *mutex, wgi *sync.WaitGroup, cntPtr *int32, t *testing.T) {
			for {
				if tm.TryLock() {
					*cntPtr = *cntPtr + 1
					tm.Unlock()
					wgi.Done()
					break
				} else {
					runtime.Gosched()
				}
			}
		}(m, wg, &cnt, t)
	}
	wg.Wait()
	//fmt.Println("count=", cnt)
	if cnt != 1000 {
		t.Error("count error concurrency")
	}
}

func Test_Mutex_TryLock_Concurrent(t *testing.T) {
	utTriableMutexConcurrent(t)
}

func Benchmark_Mutex_TryLock(b *testing.B) {
	for n := 0; n < b.N; n++ {
		utTriableMutexConcurrent(nil)
	}
}
