package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	conf.Sentinel.Stat.System.CollectIntervalMs = 0
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

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
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	fmt.Println("Sentinel Go hot-spot param flow control demo is running. You may see the pass/block metric in the metric log.")
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

	// The QPS of abc is about: 1500
	// The QPS of def is about: 50
}
