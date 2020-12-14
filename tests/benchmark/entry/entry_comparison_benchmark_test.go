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
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/logging"
)

var numbers = make([]int, 0)

func initNumberWith50() {
	numbers = make([]int, 0, 50)
	for i := 0; i < 50; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func initNumberWith100() {
	numbers = make([]int, 0, 100)
	for i := 0; i < 100; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func initNumberWith200() {
	numbers = make([]int, 0, 200)
	for i := 0; i < 200; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func initNumberWith500() {
	numbers = make([]int, 0, 500)
	for i := 0; i < 500; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func initNumberWith1000() {
	numbers = make([]int, 0, 1000)
	for i := 0; i < 1000; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func initNumberWith2000() {
	numbers = make([]int, 0, 2000)
	for i := 0; i < 2000; i++ {
		numbers = append(numbers, rand.Int())
	}
}
func initNumberWith4000() {
	numbers = make([]int, 0, 4000)
	for i := 0; i < 4000; i++ {
		numbers = append(numbers, rand.Int())
	}
}

func doSomething() {
	sort.Ints(numbers)
	//rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })
}

func doSomethingWithSentinel() {
	e, b := sentinel.Entry("benchmark_entry_comparison")
	if b != nil {
		fmt.Println("Blocked")
	} else {
		doSomething()
		e.Exit()
	}
}

func Benchmark_SlotChain_Full_Global(b *testing.B) {
	initNumberWith50()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, b := sentinel.Entry("Benchmark_SlotChain_Full_Global", sentinel.WithSlotChain(sentinel.GlobalSlotChain()))
		if b != nil {
			logging.Warn("Blocked")
		} else {
			doSomething()
			e.Exit()
		}
	}
}

func Benchmark_SlotChain_Custom_Empty(b *testing.B) {
	initNumberWith50()
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlot(stat.DefaultResourceNodePrepareSlot)
	sc.AddStatSlot(stat.DefaultSlot)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, b := sentinel.Entry("Benchmark_Custom_Empty_SlotChain", sentinel.WithSlotChain(sc))
		if b != nil {
			logging.Warn("Blocked")
		} else {
			doSomething()
			e.Exit()
		}
	}
}

func Benchmark_Single_Directly_50(b *testing.B) {
	initNumberWith50()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_50(b *testing.B) {
	initNumberWith50()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_100(b *testing.B) {
	initNumberWith100()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_100(b *testing.B) {
	initNumberWith100()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_200(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_200(b *testing.B) {
	initNumberWith200()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_500(b *testing.B) {
	initNumberWith500()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_500(b *testing.B) {
	initNumberWith500()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_1000(b *testing.B) {
	initNumberWith1000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_1000(b *testing.B) {
	initNumberWith1000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_2000(b *testing.B) {
	initNumberWith2000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_2000(b *testing.B) {
	initNumberWith2000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}

func Benchmark_Single_Directly_4000(b *testing.B) {
	initNumberWith4000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomething()
	}
}
func Benchmark_Single_StatEntry_4000(b *testing.B) {
	initNumberWith4000()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doSomethingWithSentinel()
	}
}
