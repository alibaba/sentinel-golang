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

package benchmark

import (
	"log"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
)

func InitSentinel() {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0
	conf.Sentinel.Stat.System.CollectIntervalMs = 0
	conf.Sentinel.Stat.System.CollectMemoryIntervalMs = 0
	conf.Sentinel.Stat.System.CollectCpuIntervalMs = 0
	conf.Sentinel.Stat.System.CollectLoadIntervalMs = 0
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}
