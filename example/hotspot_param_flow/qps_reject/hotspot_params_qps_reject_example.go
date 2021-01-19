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
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"

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
	rand.Seed(time.Now().UnixNano())
	testKey := "testKey"

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:        "abc",
			MetricType:      hotspot.QPS,
			ControlBehavior: hotspot.Reject,
			ParamIndex:      1,
			Threshold:       50,
			BurstCount:      0,
			DurationInSec:   1,
		},
		{
			Resource:        "def",
			MetricType:      hotspot.QPS,
			ControlBehavior: hotspot.Reject,
			ParamIndex:      2,
			Threshold:       50,
			BurstCount:      0,
			DurationInSec:   1,
		},
		{
			Resource:        "efg",
			MetricType:      hotspot.QPS,
			ControlBehavior: hotspot.Reject,
			ParamKey:        testKey,
			Threshold:       50,
			BurstCount:      0,
			DurationInSec:   1,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}
	for _, resource := range []string{"abc", "def", "efg"} {
		go func(name string) {
			node := stat.GetOrCreateResourceNode(name, base.ResTypeCommon)
			for {
				logging.Info("[HotSpot QPS] "+name,
					"pass", node.GetQPS(base.MetricEventPass),
					"block", node.GetQPS(base.MetricEventBlock),
					"complete", node.GetQPS(base.MetricEventComplete),
					"error", node.GetQPS(base.MetricEventError),
					"rt", node.GetQPS(base.MetricEventRt),
					//"\n total", node.GetQPS(base.MetricEventTotal),
				)
				time.Sleep(time.Duration(1000) * time.Millisecond)
			}
		}(resource)
	}

	logging.Info("[HotSpot Reject] Sentinel Go hot-spot param flow control demo is running. You may see the pass/block metric in the metric log.")
	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithArgs(true, rand.Uint32()%30, "sentinel"))
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
	go func() {
		for {
			e, b := sentinel.Entry("def", sentinel.WithArgs(false, 9, "ahas", fooStruct{rand.Int63() % 5}))
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

	for {
		val := fmt.Sprintf("test%v", rand.Int31()%10)
		e, b := sentinel.Entry("efg",
			sentinel.WithAttachments(map[interface{}]interface{}{
				testKey: val,
			}))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
		} else {
			// Passed, wrap the logic here.
			time.Sleep(time.Duration(rand.Uint64()%2) * time.Millisecond)
			// Be sure the entry is exited finally.
			e.Exit()
		}
	}

	// The QPS of abc is about: 1500
	// The QPS of def is about: 50
	// The QPS of efg is about: 500
}
