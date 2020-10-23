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

	// the max concurrency is 5
	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:      "abc",
			MetricType:    hotspot.Concurrency,
			ParamIndex:    0,
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
		e, b := sentinel.Entry("abc", sentinel.WithArgs(false, uint32(9)))
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
