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
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	// timespan of per slot
	BucketLengthInMs uint32 = 500
	// the number of slots
	SampleCount uint32 = 20
	// interval(ms) of sliding window, 10s
	IntervalInMs uint32 = 10 * 1000
)

func Test_bucketWrapper_Size(t *testing.T) {
	ww := &BucketWrap{
		BucketStart: util.CurrentTimeMillis(),
		Value:       atomic.Value{},
	}
	if unsafe.Sizeof(*ww) != 24 {
		t.Errorf("the size of BucketWrap is not equal 24.\n")
	}
	if unsafe.Sizeof(ww) != 8 {
		t.Errorf("the size of BucketWrap pointer is not equal 8.\n")
	}
}

// mock ArrayMock and implement BucketGenerator
type leapArrayMock struct {
	mock.Mock
}

func (bla *leapArrayMock) NewEmptyBucket() interface{} {
	return new(int64)
}

func (bla *leapArrayMock) ResetBucketTo(ww *BucketWrap, startTime uint64) *BucketWrap {
	ww.BucketStart = startTime
	ww.Value.Store(new(int64))
	return ww
}

func Test_leapArray_calculateTimeIdx_normal(t *testing.T) {
	type fields struct {
		bucketLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *AtomicBucketWrapArray
	}
	type args struct {
		timeMillis uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Test_leapArray_calculateTimeIdx_normal",
			fields: fields{
				bucketLengthInMs: BucketLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            NewAtomicBucketWrapArray(int(SampleCount), BucketLengthInMs, &leapArrayMock{}),
			},
			args: args{
				timeMillis: 1576296044907,
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LeapArray{
				bucketLengthInMs: tt.fields.bucketLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       mutex{},
			}
			if got := la.calculateTimeIdx(tt.args.timeMillis); got != tt.want {
				t.Errorf("LeapArray.calculateTimeIdx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateStartTime_normal(t *testing.T) {
	type fields struct {
	}
	type args struct {
		timeMillis       uint64
		bucketLengthInMs uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name:   "Test_calculateStartTime_normal",
			fields: fields{},
			args: args{
				timeMillis:       1576296044907,
				bucketLengthInMs: BucketLengthInMs,
			},
			want: 1576296044500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStartTime(tt.args.timeMillis, tt.args.bucketLengthInMs); got != tt.want {
				t.Errorf("LeapArray.calculateStartTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_leapArray_BucketStartCheck_normal(t *testing.T) {
	now := uint64(1596199310000)
	la := &LeapArray{
		bucketLengthInMs: BucketLengthInMs,
		sampleCount:      SampleCount,
		intervalInMs:     IntervalInMs,
		array:            NewAtomicBucketWrapArrayWithTime(int(SampleCount), BucketLengthInMs, now, &leapArrayMock{}),
		updateLock:       mutex{},
	}
	got, err := la.currentBucketOfTime(now+801, new(leapArrayMock))
	if err != nil {
		t.Errorf("LeapArray.currentBucketOfTime() error = %v\n", err)
		return
	}
	if got.BucketStart != now+500 {
		t.Errorf("BucketStart = %v, want %v", got.BucketStart, now+500)
	}
	if !reflect.DeepEqual(got, la.array.get(1)) {
		t.Errorf("LeapArray.currentBucketOfTime() = %v, want %v", got, la.array.get(1))
	}
}

func Test_leapArray_valuesWithTime_normal(t *testing.T) {
	type fields struct {
		bucketLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *AtomicBucketWrapArray
	}
	type args struct {
		timeMillis uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BucketWrap
		wantErr bool
	}{
		{
			name: "Test_leapArray_valuesWithTime_normal",
			fields: fields{
				bucketLengthInMs: BucketLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            NewAtomicBucketWrapArrayWithTime(int(SampleCount), BucketLengthInMs, uint64(1596199310000), &leapArrayMock{}),
			},
			args: args{
				timeMillis: 1576296049907,
			},
			want:    nil,
			wantErr: false,
		},
	}
	// override start time
	start := uint64(1576296040000)
	for idx := 0; idx < tests[0].fields.array.length; idx++ {
		ww := tests[0].fields.array.get(idx)
		ww.BucketStart = start
		start += 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LeapArray{
				bucketLengthInMs: tt.fields.bucketLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       mutex{},
			}
			got := la.valuesWithTime(tt.args.timeMillis)
			for _, g := range got {
				find := false
				for i := 0; i < tests[0].fields.array.length; i++ {
					w := tests[0].fields.array.get(i)
					if w.BucketStart == g.BucketStart {
						find = true
						break
					}
				}
				if !find {
					t.Errorf("LeapArray.valuesWithTime() fail")
				}
			}
		})
	}
}

func Test_leapArray_isBucketDeprecated_normal(t *testing.T) {
	type fields struct {
		bucketLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *AtomicBucketWrapArray
	}
	type args struct {
		startTime uint64
		ww        *BucketWrap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Test_leapArray_isBucketDeprecated_normal",
			fields: fields{
				bucketLengthInMs: BucketLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            NewAtomicBucketWrapArrayWithTime(int(SampleCount), BucketLengthInMs, uint64(1596199310000), &leapArrayMock{}),
			},
			args: args{
				startTime: 1576296044907,
				ww: &BucketWrap{
					BucketStart: 1576296004907,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LeapArray{
				bucketLengthInMs: tt.fields.bucketLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       mutex{},
			}
			if got := la.isBucketDeprecated(tt.args.startTime, tt.args.ww); got != tt.want {
				t.Errorf("LeapArray.isBucketDeprecated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLeapArray(t *testing.T) {
	t.Run("TestNewLeapArray_Normal", func(t *testing.T) {
		_, err := NewLeapArray(SampleCount, IntervalInMs, &leapArrayMock{})
		assert.Nil(t, err)
	})

	t.Run("TestNewLeapArray_Generator_Nil", func(t *testing.T) {
		leapArray, err := NewLeapArray(SampleCount, IntervalInMs, nil)
		assert.Nil(t, leapArray)
		assert.Error(t, err, "Invalid parameters, BucketGenerator is nil")
	})

	t.Run("TestNewLeapArray_Invalid_Parameters", func(t *testing.T) {
		leapArray, err := NewLeapArray(30, IntervalInMs, nil)
		assert.Nil(t, leapArray)
		assert.Error(t, err, "Invalid parameters, intervalInMs is 10000, sampleCount is 30")
	})
	t.Run("TestNewLeapArray_Invalid_Parameters_sampleCount0", func(t *testing.T) {
		leapArray, err := NewLeapArray(0, IntervalInMs, nil)
		assert.Nil(t, leapArray)
		assert.Error(t, err, "Invalid parameters, intervalInMs is 10000, sampleCount is 0")
	})
}
