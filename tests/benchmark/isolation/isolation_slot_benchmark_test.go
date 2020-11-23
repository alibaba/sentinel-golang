package isolation

import (
	"log"
	"math"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/tests/benchmark"
)

func init() {
	benchmark.InitSentinel()
	rule := &isolation.Rule{
		Resource:   "abc",
		MetricType: isolation.Concurrency,
		Threshold:  math.MaxInt32,
	}
	if _, err := isolation.LoadRules([]*isolation.Rule{rule}); err != nil {
		panic(err)
	}
}

func doCheck() {
	if se, err := sentinel.Entry("abc"); err == nil {
		se.Exit()
	} else {
		log.Fatalf("Block err: %s", err.Error())
	}
}

func Benchmark_IsolationSlotCheck_Loop(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doCheck()
	}
}

func Benchmark_IsolationSlotCheck_Concurrency4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_IsolationSlotCheck_Concurrency8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_IsolationSlotCheck_Concurrency16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_IsolationSlotCheck_Concurrency32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}
