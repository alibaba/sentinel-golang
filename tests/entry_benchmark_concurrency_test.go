package tests

import (
	"log"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
)

func Benchmark_Directly_Concurrency4(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomething()
		}
	})
}
func Benchmark_StatEntry_Concurrency4(b *testing.B) {
	initNumberWith200()
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinel()
		}
	})
}

func Benchmark_Directly_Concurrency8(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomething()
		}
	})
}
func Benchmark_StatEntry_Concurrency8(b *testing.B) {
	initNumberWith200()
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinel()
		}
	})
}

func Benchmark_Directly_Concurrency16(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomething()
		}
	})
}
func Benchmark_StatEntry_Concurrency16(b *testing.B) {
	initNumberWith200()
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinel()
		}
	})
}

func Benchmark_Directly_Concurrency32(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomething()
		}
	})
}
func Benchmark_StatEntry_Concurrency32(b *testing.B) {
	initNumberWith200()
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinel()
		}
	})
}
