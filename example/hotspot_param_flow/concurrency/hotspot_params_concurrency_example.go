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

package main

import (
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/logging"
)

func main() {
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
	testKey := "testKey"

	// the max concurrency is 8
	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:      "abc",
			MetricType:    hotspot.Concurrency,
			ParamIndex:    0,
			ParamKey:      testKey,
			Threshold:     8,
			DurationInSec: 0,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	go func() {
		node := stat.GetOrCreateResourceNode("abc", base.ResTypeCommon)
		for {
			logging.Info("[HotSpot Concurrency] currentConcurrency", "currentConcurrency", node.CurrentConcurrency())
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}()

	logging.Info("[HotSpot Concurrency] Sentinel Go hot-spot param flow control demo is running. You may see the pass/block metric in the metric log.")
	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithArgs(true, rand.Uint32()%30))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}

	for {
		e, b := sentinel.Entry("abc",
			sentinel.WithArgs(false, uint32(9),
				sentinel.WithAttachments(map[interface{}]interface{}{
					testKey: rand.Uint64() % 10,
				}),
			))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
		} else {
			// Passed, wrap the logic here.
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
			// Be sure the entry is exited finally.
			e.Exit()
		}
	}
}
