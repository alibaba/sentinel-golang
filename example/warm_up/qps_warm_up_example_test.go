package main

import (
	"log"
	"testing"

	"github.com/alibaba/sentinel-golang/core/flow"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func Benchmark_qps(b *testing.B) {
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "some-test",
			MetricType:      flow.QPS,
			Count:           100,
			ControlBehavior: flow.WarmUp,
			WarmUpPeriodSec: 10,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}
	for i := 0; i < b.N; i++ {
		sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
	}
}
