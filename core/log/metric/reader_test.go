package metric

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
)

func Test_readLine(t *testing.T) {

	file, _ := openFileAndSeekTo("../../../tests/testdata/extension/SystemRule.json", 100)
	defer file.Close()

	bufReader := bufio.NewReaderSize(file, 8192)

	t.Run("test1", func(t *testing.T) {
		got, err := readLine(bufReader)
		if err != nil {
			t.Errorf("readLine() error = %v", err)
			return
		}
		fmt.Println(got)
		if got != "ype\": 0," {
			t.Errorf("readLine() = %v", got)
		}
	})

}

func Test_getLatestSecond(t *testing.T) {
	type args struct {
		items []*base.MetricItem
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"test1", args{nil}, 0},
		// {"test2", args{[]*base.MetricItem{}}, 0}, //
		{"test3", args{[]*base.MetricItem{&base.MetricItem{}}}, 0},
		{
			"test4",
			args{
				[]*base.MetricItem{
					&base.MetricItem{Timestamp: 1582421778887},
				},
			},
			1582421778,
		},
		{
			"test5",
			args{
				[]*base.MetricItem{
					&base.MetricItem{Timestamp: 1582421778887},
					&base.MetricItem{Timestamp: 1577808000000},
				},
			},
			1577808000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestSecond(tt.args.items); got != tt.want {
				t.Errorf("getLatestSecond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_openFileAndSeekTo(t *testing.T) {

	filename := "../../../tests/testdata/metric/app1-metrics.log.2020-02-14"

	type args struct {
		filename string
		offset   uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test1",
			args{
				filename,
				1,
			},
			false,
		}, {
			"test2",
			args{
				filename,
				100000000,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := openFileAndSeekTo(tt.args.filename, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("openFileAndSeekTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
