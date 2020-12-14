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

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/logging"
)

type fooStruct struct {
	n int64
}

func main() {
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:          "abc",
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        1,
			Threshold:         1000,
			MaxQueueingTimeMs: 5,
			DurationInSec:     1,
		},
		{
			Resource:          "def",
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        1,
			Threshold:         1000,
			MaxQueueingTimeMs: 5,
			DurationInSec:     1,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	logging.Info("[HotSpot Throttling] Sentinel Go hot-spot param flow control demo is running. You may see the pass/block metric in the metric log.")
	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithArgs(true, rand.Uint32()%30, "sentinel"))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
				} else {
					// Passed, wrap the logic here.
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}

	for {
		e, b := sentinel.Entry("def", sentinel.WithArgs(false, uint32(9), "ahas", fooStruct{rand.Int63() % 5}))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
		} else {
			// Passed, wrap the logic here.
			// Be sure the entry is exited finally.
			e.Exit()
		}
	}
	// The QPS of abc is about: 15000
	// The QPS of def is about: 950
}
