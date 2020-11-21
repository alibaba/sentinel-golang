package flow

import (
	"log"
	"math"
	"testing"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
)

func doCheck() {
	if se, err := api.Entry("abc"); err == nil {
		se.Exit()
	} else {
		log.Fatalf("Block err: %s", err.Error())
	}
}

func loadDirectRejectRule() {
	rule := &flow.Rule{
		Resource:               "abc",
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
		Threshold:              math.MaxFloat64,
		StatIntervalInMs:       1000,
		RelationStrategy:       flow.CurrentResource,
	}
	flow.LoadRules([]*flow.Rule{rule})
}

func loadWarmUpRejectRule() {
	rule := &flow.Rule{
		Resource:               "abc",
		TokenCalculateStrategy: flow.WarmUp,
		ControlBehavior:        flow.Reject,
		Threshold:              math.MaxFloat64,
		WarmUpPeriodSec:        10,
		WarmUpColdFactor:       3,
		StatIntervalInMs:       1000,
	}
	flow.LoadRules([]*flow.Rule{rule})
}

func Benchmark_DirectReject_SlotCheck_4(b *testing.B) {
	loadDirectRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_DirectReject_SlotCheck_8(b *testing.B) {
	loadDirectRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_DirectReject_SlotCheck_16(b *testing.B) {
	loadDirectRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_DirectReject_SlotCheck_32(b *testing.B) {
	loadDirectRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_WarmUpReject_SlotCheck_4(b *testing.B) {
	loadWarmUpRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_WarmUpReject_SlotCheck_8(b *testing.B) {
	loadWarmUpRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_WarmUpReject_SlotCheck_16(b *testing.B) {
	loadWarmUpRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_WarmUpReject_SlotCheck_32(b *testing.B) {
	loadWarmUpRejectRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}
