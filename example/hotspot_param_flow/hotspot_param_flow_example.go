package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/util"
)

type fooStruct struct {
	n int64
}

func main() {
	m := make([]hotspot.SpecificValue, 1)
	m[0] = hotspot.SpecificValue{
		ValKind:   hotspot.KindInt,
		ValStr:    "9",
		Threshold: 0,
	}
	var Resource = "test"

	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:          Resource,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        1,
			Threshold:         50,
			MaxQueueingTimeMs: 0,
			BurstCount:        0,
			DurationInSec:     1,
			SpecificItems:     m,
		},
		{
			Resource:          Resource,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        2,
			Threshold:         50,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     make([]hotspot.SpecificValue, 0),
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
				e, b := sentinel.Entry(Resource, sentinel.WithArgs(true, rand.Uint32()%30, "sentinel", fooStruct{rand.Int63() % 5}))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%50) * time.Millisecond)
					fmt.Println(util.CurrentTimeMillis(), b.Error())
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%50) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}

	for {
		e, b := sentinel.Entry(Resource, sentinel.WithArgs(false, uint32(9), "ahas", fooStruct{rand.Int63() % 5}))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
			time.Sleep(time.Duration(rand.Uint64()%50) * time.Millisecond)
		} else {
			// Passed, wrap the logic here.
			fmt.Println(util.CurrentTimeMillis(), "passed")
			time.Sleep(time.Duration(rand.Uint64()%50) * time.Millisecond)

			// Be sure the entry is exited finally.
			e.Exit()
		}
	}
}
