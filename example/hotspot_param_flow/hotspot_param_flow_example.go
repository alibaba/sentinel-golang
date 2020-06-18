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
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Id:                "1",
			Resource:          Resource,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        1,
			Threshold:         50,
			MaxQueueingTimeMs: 0,
			BurstCount:        0,
			DurationInSec:     1,
			SpecificItems: map[hotspot.SpecificValue]int64{
				{ValKind: hotspot.KindInt, ValStr: "9"}: 0,
			},
		},
		{
			Id:                "2",
			Resource:          Resource,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        2,
			Threshold:         50,
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

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry(Resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithSlotChain(sc), sentinel.WithArgs(true, rand.Uint32()%30, "sentinel", uuid.New().String()))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					//fmt.Println(util.CurrentTimeMillis(), "blocked")
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					//time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}

	for {
		e, b := sentinel.Entry(Resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithSlotChain(sc), sentinel.WithArgs(false, uint32(9), "ahas", uuid.New().String()))
		if b != nil {
			// Blocked. We could get the block reason from the BlockError.
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
			fmt.Println(util.CurrentTimeMillis(), "blocked")
		} else {
			// Passed, wrap the logic here.
			fmt.Println(util.CurrentTimeMillis(), "passed")
			time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

			// Be sure the entry is exited finally.
			e.Exit()
		}
	}
}
