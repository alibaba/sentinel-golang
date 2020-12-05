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

package memory

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func doSomethingWithSentinelWithResource(res string) {
	e, b := sentinel.Entry(res, sentinel.WithTrafficType(base.Inbound))
	if b != nil {
		fmt.Println("Blocked")
	} else {
		e.Exit()
	}
}

func TestMemory_Single(t *testing.T) {
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	i := 0
	for {
		i++
		for i := 0; i < 6000; i++ {
			doSomethingWithSentinelWithResource(strconv.Itoa(i))
		}
		if i == 1 {
			break
		}
	}
}

func TestMemory_Concurrency(t *testing.T) {
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	// prepare statistic structure
	for i := 0; i < 6000; i++ {
		doSomethingWithSentinelWithResource(strconv.Itoa(i))
	}

	wg := &sync.WaitGroup{}
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			c := 0
			for {
				c++
				for j := 0; j < 6000; j++ {
					doSomethingWithSentinelWithResource(strconv.Itoa(j))
				}
				if c == 1 {
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
