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
