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

package base

import (
	"math/rand"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
)

func BenchmarkBucketLeapArray_AddCount_Concurrency1(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.AddCount(base.MetricEventPass, 1)
		}
	})
}

func BenchmarkBucketLeapArray_AddCount_Concurrency10(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(10)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.AddCount(base.MetricEventPass, 1)
		}
	})
}

func BenchmarkBucketLeapArray_AddCount_Concurrency100(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(100)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.AddCount(base.MetricEventPass, 1)
		}
	})
}

func BenchmarkBucketLeapArray_AddCount_Concurrency1000(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(1000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.AddCount(base.MetricEventPass, 1)
		}
	})
}

func BenchmarkBucketLeapArray_Count_Concurrency1(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(1)
	b.ResetTimer()
	rand.Seed(util.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Int()%100 >= 80 {
				a.AddCount(base.MetricEventPass, 1)
			} else {
				_ = a.Count(base.MetricEventPass)
			}
		}
	})
}

func BenchmarkBucketLeapArray_Count_Concurrency10(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(10)
	b.ResetTimer()
	rand.Seed(util.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Int()%100 >= 80 {
				a.AddCount(base.MetricEventPass, 1)
			} else {
				_ = a.Count(base.MetricEventPass)
			}
		}
	})
}
func BenchmarkBucketLeapArray_Count_Concurrency100(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(100)
	b.ResetTimer()
	rand.Seed(util.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Int()%100 >= 80 {
				a.AddCount(base.MetricEventPass, 1)
			} else {
				_ = a.Count(base.MetricEventPass)
			}
		}
	})
}

func BenchmarkBucketLeapArray_Count_Concurrency1000(b *testing.B) {
	a := NewBucketLeapArray(base.DefaultSampleCountTotal, base.DefaultIntervalMsTotal)
	b.ReportAllocs()
	b.SetParallelism(1000)
	b.ResetTimer()
	rand.Seed(util.Now().UnixNano())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Int()%100 >= 80 {
				a.AddCount(base.MetricEventPass, 1)
			} else {
				_ = a.Count(base.MetricEventPass)
			}
		}
	})
}
