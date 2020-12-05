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
)

//Time zone should be considered for time related operations
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

func BenchmarkCurrentTimeInMs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CurrentTimeMillis()
	}
}

func BenchmarkCurrentTimeInMsWithTicker(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CurrentTimeMillsWithTicker()
	}
}
