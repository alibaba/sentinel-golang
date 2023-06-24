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

package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Time zone should be considered for time related operations
func TestFormatTimeMillis(t *testing.T) {
	type args struct {
		ts uint64
	}
	_, offset := time.Now().Zone()
	offsetMs := uint64(offset * 1000)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test1",
			args{1582421778887 - offsetMs},
			"2020-02-23 01:36:18", //"2020-02-23 09:36:18",
		}, {
			"test2",
			args{1577808000000 - offsetMs},
			"2019-12-31 16:00:00", //"2020-01-01 00:00:00",
		}, {
			"test3",
			args{1582423015000 - offsetMs},
			"2020-02-23 01:56:55", //"2020-02-23 09:56:55",
		}, {
			"test4",
			args{1564382218000 - offsetMs},
			"2019-07-29 06:36:58", //"2019-07-29 14:36:58",
		}, {
			"test5",
			args{1582427442295 - offsetMs},
			"2020-02-23 03:10:42", //"2020-02-23 11:10:42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTimeMillis(tt.args.ts); got != tt.want {
				t.Errorf("%s FormatTimeMillis() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	type args struct {
		tsMillis uint64
	}
	_, offset := time.Now().Zone()
	offsetMs := uint64(offset * 1000)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test1",
			args{1564382218000 - offsetMs},
			"2019-07-29",
		}, {
			"test2",
			args{1577808000000 - offsetMs},
			"2019-12-31", //"2020-01-01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDate(tt.args.tsMillis); got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrentTimeMillis(t *testing.T) {
	got := CurrentTimeMillis()
	fmt.Printf("CurrentTimeMillis() %d", got)
	fmt.Println(FormatTimeMillis(got))
}

func TestCurrentTimeNano(t *testing.T) {
	got := CurrentTimeNano()
	fmt.Println(got)
}

func TestMockClock(t *testing.T) {
	clock := NewMockClock()

	now := clock.Now()
	assert.Equal(t, clock.CurrentTimeNano(), uint64(now.UnixNano()))
	assert.Equal(t, clock.CurrentTimeMillis(), uint64(now.UnixNano()/1e6))

	last := now
	d := 23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond + 999*time.Microsecond
	clock.Sleep(d)
	assert.Equal(t, last.Add(d), clock.Now())
}

func TestMockTicker(t *testing.T) {
	SetClock(NewMockClock())
	defer SetClock(NewRealClock())

	d := 23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond + 999*time.Microsecond

	ticker := NewMockTicker(d)
	defer ticker.Stop()

	ticked := 0

	ticker.check()
	select {
	case <-ticker.C():
		ticked++
	default:
	}
	assert.Equal(t, ticked, 0)

	Sleep(d - 1)
	ticker.check()
	select {
	case <-ticker.C():
		ticked++
	default:
	}
	assert.Equal(t, ticked, 0)

	Sleep(1)
	ticker.check()
	select {
	case <-ticker.C():
		ticked++
	default:
	}
	assert.Equal(t, ticked, 1)

	Sleep(d + 1)
	ticker.check()
	select {
	case <-ticker.C():
		ticked++
	default:
	}
	assert.Equal(t, ticked, 2)

	Sleep(d - 1)
	ticker.check()
	select {
	case <-ticker.C():
		ticked++
	default:
	}
	assert.Equal(t, ticked, 3)
}

func BenchmarkCurrentTimeInMs(b *testing.B) {
	StartTimeTicker()
	currentTimeMillisDirect := func() uint64 {
		tickerNow := CurrentTimeMillsWithTicker()
		if tickerNow > uint64(0) {
			return tickerNow
		}
		return uint64(time.Now().UnixNano()) / UnixTimeUnitOffset
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.Run("CurrentTimeMillis", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CurrentTimeMillis()
		}
	})
	b.Run("CurrentTimeMillisDirect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			currentTimeMillisDirect()
		}
	})
}

func BenchmarkCurrentTimeInMsWithTicker(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CurrentTimeMillsWithTicker()
	}
}

func BenchmarkNow(b *testing.B) {
	b.Run("util.Now", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Now()
		}
	})
	b.Run("time.Now", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			time.Now()
		}
	})
}
