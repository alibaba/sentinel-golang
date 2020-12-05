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

package misc

import (
	"log"
	"testing"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/logging"
)

func Test_Flow_Slot_LoadRules(t *testing.T) {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "abc",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              10,
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	go func() {
		time.Sleep(time.Second * 2)
		_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{
			// Statistic time span=10s, recoveryTimeout=3s, slowRtUpperBound=50ms, maxSlowRequestRatio=50%
			{
				Resource:         "abc",
				Strategy:         circuitbreaker.SlowRequestRatio,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 10,
				StatIntervalMs:   5000,
				MaxAllowedRtMs:   50,
				Threshold:        0.5,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}()

	ch := time.NewTicker(time.Second * 3).C

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
				} else {
					// Passed, wrap the logic here.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}
