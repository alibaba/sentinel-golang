package circuitbreaker

import (
	"log"
	"testing"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

func doCheck() {
	if se, err := api.Entry("abc"); err == nil {
		se.Exit()
	} else {
		log.Fatalf("Block err: %s", err.Error())
	}
}

func loadSlowRequestRatioRule() {
	rule := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   20,
		Threshold:        0.1,
	}
	circuitbreaker.LoadRules([]*circuitbreaker.Rule{rule})
}

func loadErrorRatioRule() {
	rule := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.ErrorRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        0.3,
	}
	circuitbreaker.LoadRules([]*circuitbreaker.Rule{rule})
}

func loadErrorCountRule() {
	rule := &circuitbreaker.Rule{
		Resource:         "abc",
		Strategy:         circuitbreaker.ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10.0,
	}
	circuitbreaker.LoadRules([]*circuitbreaker.Rule{rule})
}

func Benchmark_SlowRequestRatio_SlotCheck_4(b *testing.B) {
	loadSlowRequestRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_8(b *testing.B) {
	loadSlowRequestRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_16(b *testing.B) {
	loadSlowRequestRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_32(b *testing.B) {
	loadSlowRequestRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_4(b *testing.B) {
	loadErrorRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_8(b *testing.B) {
	loadErrorRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_16(b *testing.B) {
	loadErrorRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_32(b *testing.B) {
	loadErrorRatioRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_4(b *testing.B) {
	loadErrorCountRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_8(b *testing.B) {
	loadErrorCountRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_16(b *testing.B) {
	loadErrorCountRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_32(b *testing.B) {
	loadErrorCountRule()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck()
		}
	})
}
