package hotspot

import (
	"log"
	"math"
	"testing"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/hotspot"
)

func doCheck() {
	if se, err := api.Entry("abc"); err == nil {
		se.Exit()
	} else {
		log.Fatalf("Block err: %s", err.Error())
	}
}

func initConcurrencyRule() {
	_, err := hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:      "abc",
			MetricType:    hotspot.Concurrency,
			ParamIndex:    0,
			Threshold:     math.MaxInt64,
			DurationInSec: 0,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
}

func initQPSRejectRule() {
	_, err := hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:        "abc",
			MetricType:      hotspot.QPS,
			ControlBehavior: hotspot.Reject,
			ParamIndex:      0,
			Threshold:       math.MaxInt64,
			BurstCount:      0,
			DurationInSec:   1,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
}

func initQPSThrottlingRule() {
	_, err := hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:          "abc",
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Throttling,
			ParamIndex:        0,
			Threshold:         math.MaxInt64,
			BurstCount:        0,
			DurationInSec:     1,
			MaxQueueingTimeMs: 0,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
}

func Benchmark_Concurrency_Concurrency4(b *testing.B) {
	initConcurrencyRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_Concurrency_Concurrency8(b *testing.B) {
	initConcurrencyRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_Concurrency_Concurrency16(b *testing.B) {
	initConcurrencyRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Reject_Concurrency4(b *testing.B) {
	initQPSRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Reject_Concurrency8(b *testing.B) {
	initQPSRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Reject_Concurrency16(b *testing.B) {
	initQPSRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Throttling_Concurrency4(b *testing.B) {
	initQPSThrottlingRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Throttling_Concurrency8(b *testing.B) {
	initQPSThrottlingRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_QPS_Throttling_Concurrency16(b *testing.B) {
	initQPSThrottlingRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}
