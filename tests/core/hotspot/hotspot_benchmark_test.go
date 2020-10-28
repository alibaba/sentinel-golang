package hotspot

import (
	"log"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/stat"
)

var (
	ctx *base.EntryContext
)

func init() {
	ctx = base.NewEmptyEntryContext()
	ctx.Resource = base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound)
	ctx.Input = &base.SentinelInput{
		BatchCount:  1,
		Flag:        0,
		Args:        make([]interface{}, 0),
		Attachments: make(map[interface{}]interface{}),
	}
	ctx.Input.Args = append(ctx.Input.Args, "test")
	ctx.StatNode = stat.GetOrCreateResourceNode("abc", base.ResTypeCommon)
}

func initConcurrencyRule() {
	_, err := hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:      "abc",
			MetricType:    hotspot.Concurrency,
			ParamIndex:    0,
			Threshold:     100,
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
			Threshold:       100,
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
			Threshold:         100,
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
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
			hotspot.DefaultSlot.Check(ctx)
		}
	})
}
