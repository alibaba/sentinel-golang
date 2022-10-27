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
	"sync/atomic"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	_ "github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/pkg/datasource/opensergo"
	_ "github.com/alibaba/sentinel-golang/pkg/datasource/opensergo"
	"github.com/alibaba/sentinel-golang/util"
)

const (
	host string = "127.0.0.1"
	port uint32 = 10246

	namespace string = "default"
	app       string = "foo-app"
)

type Counter struct {
	pass  *int64
	block *int64
	total *int64
}

func main() {
	openSergoDataSource, _ := opensergo.NewOpenSergoDataSource(host, port, namespace, app)
	openSergoDataSource.Initialize()

	// simulate concurrency request
	simulateConcurrency()

	select {}
}

func simulateConcurrency() {
	counter := Counter{pass: new(int64), block: new(int64), total: new(int64)}

	// simulate request
	go startFlowModule(&counter)
	// print counter
	go timerTask(&counter)
}

func startFlowModule(counter *Counter) {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	if err := sentinel.InitWithConfig(conf); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("GET:/foo/1", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					atomic.AddInt64(counter.block, 1)
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					atomic.AddInt64(counter.pass, 1)
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}
				atomic.AddInt64(counter.total, 1)
			}
		}()
	}
}

//statistic print
func timerTask(counter *Counter) {
	fmt.Println("begin to statistic!!!")
	var (
		oldTotal, oldPass, oldBlock int64
	)
	for {
		time.Sleep(1 * time.Second)
		globalTotal := atomic.LoadInt64(counter.total)
		oneSecondTotal := globalTotal - oldTotal
		oldTotal = globalTotal

		globalPass := atomic.LoadInt64(counter.pass)
		oneSecondPass := globalPass - oldPass
		oldPass = globalPass

		globalBlock := atomic.LoadInt64(counter.block)
		oneSecondBlock := globalBlock - oldBlock
		oldBlock = globalBlock

		fmt.Println(util.CurrentTimeMillis()/1000, "total:", oneSecondTotal, " pass:", oneSecondPass, " block:", oneSecondBlock)
	}
}
