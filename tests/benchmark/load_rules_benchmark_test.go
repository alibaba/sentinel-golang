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

package benchmark

import (
	"math/rand"
	"strconv"
	"testing"

	cb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

func Test_Size_1000_Circuit_Breaker_Rules_Update(t *testing.T) {
	logging.ResetGlobalLoggerLevel(logging.ErrorLevel)
	rs := make([]*cb.Rule, 0, 1000)
	rand.Seed(int64(util.CurrentTimeMillis()))
	intervals := []uint32{10000, 15000, 20000, 25000, 30000}
	for i := 0; i < 1000; i++ {
		retryTimeout := intervals[rand.Int()%5]
		rs = append(rs, &cb.Rule{
			Resource:         "github.com/alibaba/sentinel/test",
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   retryTimeout,
			MinRequestAmount: rand.Uint64() % 100,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   100,
			Threshold:        0.1,
		})
	}

	_, err := cb.LoadRules(rs)
	if err != nil {
		t.Errorf("error")
	}
	logging.ResetGlobalLoggerLevel(logging.InfoLevel)
}

func Benchmark_Size_1000_Circuit_Breaker_Rules_Update(b *testing.B) {
	rs := make([]*cb.Rule, 0, 1000)
	rand.Seed(int64(util.CurrentTimeMillis()))
	intervals := []uint32{10000, 15000, 20000, 25000, 30000}
	for i := 0; i < 1000; i++ {
		retryTimeout := intervals[rand.Int()%5]
		rs = append(rs, &cb.Rule{
			Resource:         "github.com/alibaba/sentinel/test" + strconv.Itoa(rand.Int()%100),
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   retryTimeout,
			MinRequestAmount: rand.Uint64() % 100,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   100,
			Threshold:        0.1,
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cb.LoadRules(rs)
		if err != nil {
			b.Errorf("error")
		}
	}
}

func Test_Size_10000_Circuit_Breaker_Rules_Update(t *testing.T) {
	logging.ResetGlobalLoggerLevel(logging.ErrorLevel)
	rs := make([]*cb.Rule, 0, 10000)
	intervals := []uint32{10000, 15000, 20000, 25000, 30000}
	for i := 0; i < 10000; i++ {
		retryTimeout := intervals[rand.Int()%5]
		rs = append(rs, &cb.Rule{
			Resource:         "github.com/alibaba/sentinel/test" + strconv.Itoa(rand.Int()%100),
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   retryTimeout,
			MinRequestAmount: rand.Uint64() % 100,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   100,
			Threshold:        0.1,
		})
	}

	_, err := cb.LoadRules(rs)
	if err != nil {
		t.Errorf("error")
	}
	logging.ResetGlobalLoggerLevel(logging.InfoLevel)
}

func Benchmark_Size_10000_Circuit_Breaker_Rules_Update(b *testing.B) {
	rs := make([]*cb.Rule, 0, 10000)
	intervals := []uint32{10000, 15000, 20000, 25000, 30000}
	for i := 0; i < 10000; i++ {
		retryTimeout := intervals[rand.Int()%5]
		rs = append(rs, &cb.Rule{
			Resource:         "github.com/alibaba/sentinel/test" + strconv.Itoa(rand.Int()%100),
			Strategy:         cb.SlowRequestRatio,
			RetryTimeoutMs:   retryTimeout,
			MinRequestAmount: rand.Uint64() % 100,
			StatIntervalMs:   10000,
			MaxAllowedRtMs:   100,
			Threshold:        0.1,
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cb.LoadRules(rs)
		if err != nil {
			b.Errorf("error")
		}
	}
}
