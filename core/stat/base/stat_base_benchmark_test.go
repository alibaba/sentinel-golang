package base

import (
	"math/rand"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
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
	rand.Seed(time.Now().UnixNano())
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
	rand.Seed(time.Now().UnixNano())
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
	rand.Seed(time.Now().UnixNano())
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
	rand.Seed(time.Now().UnixNano())
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
