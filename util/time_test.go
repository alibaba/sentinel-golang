package util

import (
	"fmt"
	"testing"
)

func TestFormatTimeMillis(t *testing.T) {
	// timeUnix := time.Now().Unix() //已知的时间戳，精确到秒
	// fmt.Println(timeUnix)
	// formatTimeStr := time.Unix(timeUnix, 0).Format("2006-01-02 15:04:05")
	// fmt.Println(formatTimeStr) //打印结果：2017-04-11 13:30:39

	type args struct {
		ts uint64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test1",
			args{1582421778887},
			"2020-02-23 09:36:18",
		}, {
			"test2",
			args{1577808000000},
			"2020-01-01 00:00:00",
		}, {
			"test3",
			args{1582423015000},
			"2020-02-23 09:56:55",
		}, {
			"test3",
			args{1582423015000},
			"2020-02-23 09:56:55",
		}, {
			"test4",
			args{1564382218000},
			"2019-07-29 14:36:58",
		}, {
			"test5",
			args{1582427442295},
			"2020-02-23 11:10:42",
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
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test1",
			args{1564382218000},
			"2019-07-29",
		}, {
			"test2",
			args{1577808000000},
			"2020-01-01",
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
	times := CurrentTimeMillis()
	fmt.Println(times)
	fmt.Println(FormatTimeMillis(times))
}
