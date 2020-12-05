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

package entry

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
)

func doSomethingWithSize() {
	numbers := make([]int, 0, 200)
	for i := 0; i < 200; i++ {
		numbers = append(numbers, rand.Int())
	}
	sort.Ints(numbers)
	//rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })
}

func doSomethingWithSentinelForConcurrency() {
	e, b := sentinel.Entry("benchmark_entry_comparison_concurrency")
	if b != nil {
		fmt.Println("Blocked")
	} else {
		doSomethingWithSize()
		e.Exit()
	}
}

func Benchmark_Concurrency_Directly_4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSize()
		}
	})
}
func Benchmark_Concurrency_StatEntry_4(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinelForConcurrency()
		}
	})
}

func Benchmark_Concurrency_Directly_8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSize()
		}
	})
}
func Benchmark_Concurrency_StatEntry_8(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinelForConcurrency()
		}
	})
}

func Benchmark_Concurrency_Directly_16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSize()
		}
	})
}
func Benchmark_Concurrency_StatEntry_16(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinelForConcurrency()
		}
	})
}

func Benchmark_Concurrency_Directly_32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSize()
		}
	})
}
func Benchmark_Concurrency_StatEntry_32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doSomethingWithSentinelForConcurrency()
		}
	})
}
