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

package metric

import (
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormMetricFileName(t *testing.T) {
	appName1 := "foo-test"
	appName2 := "foo.test"
	mf1 := FormMetricFileName(appName1, false)
	mf2 := FormMetricFileName(appName2, false)
	assert.Equal(t, "foo-test-metrics.log", mf1)
	assert.Equal(t, mf1, mf2)
	mf1Pid := FormMetricFileName(appName2, true)
	if !strings.HasSuffix(mf1Pid, strconv.Itoa(os.Getpid())) {
		t.Fatalf("Metric filename <%s> should end with the process id", mf1Pid)
	}
}

func Test_filenameMatches(t *testing.T) {
	type args struct {
		filename     string
		baseFilename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_filenameMatches",
			args: args{
				filename:     "~/logs/csp/app1-metric.log.2018-12-24.1111",
				baseFilename: "~/logs/csp/app1-metric.log",
			},
			want: true,
		},
		{
			name: "Test_filenameMatches",
			args: args{
				filename:     "~/logs/csp/app1-metric.log-2018-12-24.1111",
				baseFilename: "~/logs/csp/app1-metric.log",
			},
			want: false,
		},
		{
			name: "Test_filenameMatches",
			args: args{
				filename:     "~/logs/csp/app2-metric.log.2018-12-24.1111",
				baseFilename: "~/logs/csp/app1-metric.log",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filenameMatches(tt.args.filename, tt.args.baseFilename); got != tt.want {
				t.Errorf("filenameMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilenameComparatorNoPid(t *testing.T) {
	arr := []string{
		"metrics.log.2018-03-06",
		"metrics.log.2018-03-07",
		"metrics.log.2018-03-07.51",
		"metrics.log.2018-03-07.10",
		"metrics.log.2018-03-06.100",
	}
	expected := []string{
		"metrics.log.2018-03-06",
		"metrics.log.2018-03-06.100",
		"metrics.log.2018-03-07",
		"metrics.log.2018-03-07.10",
		"metrics.log.2018-03-07.51",
	}

	sort.Slice(arr, filenameComparator(arr))
	assert.Equal(t, expected, arr)
}

func Test_listMetricFiles(t *testing.T) {
	type args struct {
		baseDir     string
		filePattern string
	}
	var tests = []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Test_listMetricFiles",
			args: args{
				baseDir:     "../../../tests/testdata/metric",
				filePattern: "app1-metrics.log",
			},
			want: []string{
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-14",
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-14.12",
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-14.32",
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-15",
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-16",
				"../../../tests/testdata/metric/app1-metrics.log.2020-02-16.100",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listMetricFiles(tt.args.baseDir, tt.args.filePattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("listMetricFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if osType := runtime.GOOS; osType == "windows" {
				for i := 0; i < len(got); i++ {
					got[i] = strings.ReplaceAll(got[i], "\\", "/")
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listMetricFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
