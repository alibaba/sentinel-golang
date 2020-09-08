package main

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/util"

	"github.com/alibaba/sentinel-golang/core/flow"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func Benchmark_qps(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doTest()
	}
}

func doTest() {
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "some-test",
			MetricType:             flow.QPS,
			Count:                  100,
			TokenCalculateStrategy: flow.WarmUp,
			ControlBehavior:        flow.Reject,
			WarmUpPeriodSec:        10,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}
	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	time.Sleep(time.Second * 5)
}
