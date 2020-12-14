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
	"math/rand"
	"os"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/logging"
)

func main() {
	cfg := config.NewDefaultConfig()
	// for testing, logging output to console
	cfg.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(cfg)
	if err != nil {
		logging.Error(err, "fail")
		os.Exit(1)
	}
	logging.ResetGlobalLoggerLevel(logging.DebugLevel)
	ch := make(chan struct{})

	r1 := &isolation.Rule{
		Resource:   "abc",
		MetricType: isolation.Concurrency,
		Threshold:  12,
	}
	_, err = isolation.LoadRules([]*isolation.Rule{r1})
	if err != nil {
		logging.Error(err, "fail")
		os.Exit(1)
	}

	for i := 0; i < 15; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithBatchCount(1))
				if b != nil {
					logging.Info("[Isolation] Blocked", "reason", b.BlockType().String(), "rule", b.TriggeredRule(), "snapshot", b.TriggeredValue())
					time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
				} else {
					logging.Info("[Isolation] Passed")
					time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
					e.Exit()
				}
			}
		}()
	}
	<-ch
}
