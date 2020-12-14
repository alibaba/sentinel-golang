// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package circuitbreaker

import (
	"log"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/tests/benchmark"
)

const (
	resSlowRt     = "abc-slowRt"
	resErrorRatio = "abc-errorRatio"
	resErrorCount = "abc-errorCount"
)

func doCheck(res string) {
	if se, err := sentinel.Entry(res); err == nil {
		se.Exit()
	} else {
		log.Fatalf("Block err: %s", err.Error())
	}
}

func init() {
	benchmark.InitSentinel()
	rule1 := &circuitbreaker.Rule{
		Resource:         resSlowRt,
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		MaxAllowedRtMs:   100000,
		Threshold:        0.99,
	}
	rule2 := &circuitbreaker.Rule{
		Resource:         resErrorRatio,
		Strategy:         circuitbreaker.ErrorRatio,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        0.99,
	}
	rule3 := &circuitbreaker.Rule{
		Resource:         resErrorCount,
		Strategy:         circuitbreaker.ErrorCount,
		RetryTimeoutMs:   1000,
		MinRequestAmount: 5,
		StatIntervalMs:   1000,
		Threshold:        10000.0,
	}
	_, err := circuitbreaker.LoadRules([]*circuitbreaker.Rule{rule1, rule2, rule3})
	if err != nil {
		panic(err)
	}
}

func Benchmark_SlowRequestRatio_SlotCheck_4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resSlowRt)
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resSlowRt)
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resSlowRt)
		}
	})
}

func Benchmark_SlowRequestRatio_SlotCheck_32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resSlowRt)
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorRatio)
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorRatio)
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorRatio)
		}
	})
}

func Benchmark_ErrorRatio_SlotCheck_32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorRatio)
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorCount)
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorCount)
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorCount)
		}
	})
}

func Benchmark_ErrorCount_SlotCheck_32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doCheck(resErrorCount)
		}
	})
}
