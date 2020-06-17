package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/google/uuid"
)

func main() {
	var Resource = "test"

	// We should initialize Sentinel first.
	err := sentinel.Init("./hot-pramas-sentinel.yml")
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Id:                "a1",
			Resource:          Resource,
			MetricType:        hotspot.Concurrency,
			Behavior:          hotspot.Reject,
			ParamIndex:        0,
			Threshold:         100,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     make(map[hotspot.SpecificValue]int64),
		},
		{
			Id:                "a2",
			Resource:          Resource,
			MetricType:        hotspot.Concurrency,
			Behavior:          hotspot.Reject,
			ParamIndex:        1,
			Threshold:         100,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     make(map[hotspot.SpecificValue]int64),
		},
		{
			Id:                "a3",
			Resource:          Resource,
			MetricType:        hotspot.Concurrency,
			Behavior:          hotspot.Reject,
			ParamIndex:        2,
			Threshold:         100,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     make(map[hotspot.SpecificValue]int64),
		},
		{
			Id:                "a4",
			Resource:          Resource,
			MetricType:        hotspot.Concurrency,
			Behavior:          hotspot.Reject,
			ParamIndex:        3,
			Threshold:         100,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     make(map[hotspot.SpecificValue]int64),
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	sc := base.NewSlotChain()
	sc.AddStatPrepareSlotLast(&stat.StatNodePrepareSlot{})
	sc.AddRuleCheckSlotLast(&system.SystemAdaptiveSlot{})
	sc.AddRuleCheckSlotLast(&flow.FlowSlot{})
	sc.AddRuleCheckSlotLast(&hotspot.FreqPramsTrafficSlot{})
	sc.AddStatSlotLast(&stat.StatisticSlot{})
	sc.AddStatSlotLast(&hotspot.ConcurrencyStatSlot{})

	for i := 0; i < 100; i++ {
		go func() {
			for {
				e, b := sentinel.Entry(Resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithSlotChain(sc), sentinel.WithArgs(true, rand.Int()%3000, uuid.New().String(), uuid.New().String()))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					fmt.Println(util.CurrentTimeMillis(), " blocked")
				} else if e == nil && b == nil {
					fmt.Println("e is ni")
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), " passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}

	for {
		e, b := sentinel.Entry(Resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithSlotChain(sc), sentinel.WithArgs(true, rand.Int()%3000, uuid.New().String(), uuid.New().String()))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
			fmt.Println(util.CurrentTimeMillis(), " blocked")
		} else if e == nil && b == nil {
			fmt.Println("e is ni")
		} else {
			// Passed, wrap the logic here.
			fmt.Println(util.CurrentTimeMillis(), " passed")
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

			// Be sure the entry is exited finally.
			e.Exit()
		}

	}
}
