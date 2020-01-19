package base

import (
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/sentinel-group/sentinel-golang/util"
	"github.com/stretchr/testify/mock"
)

const (
	// timespan of per slot
	WindowLengthInMs uint32 = 500
	// the number of slots
	SampleCount uint32 = 20
	// interval(ms) of sliding window, 10s
	IntervalInMs uint32 = 10 * 1000
)

func Test_windowWrapper_Size(t *testing.T) {
	type Obj struct {
		a1 int32 // 4bytes
		a2 int32
		a3 int32
		a4 int32
		a5 int32
		a6 int32
		a7 int32
		a8 int32
	}
	ww := &windowWrap{
		windowStart: util.CurrentTimeMillis(),
		value:       atomic.Value{},
	}
	if unsafe.Sizeof(*ww) != 24 {
		t.Errorf("the size of windowWrap is not equal 20.\n")
	}
	if unsafe.Sizeof(ww) != 8 {
		t.Errorf("the size of windowWrap is not equal 20.\n")
	}
}

//type metricBucketMock struct {
//	mock.Mock
//}

// mock ArrayMock and implement bucketGenerator
type leapArrayMock struct {
	mock.Mock
}

func (bla *leapArrayMock) newEmptyBucket() interface{} {
	return new(int64)
}

func (bla *leapArrayMock) resetWindowTo(ww *windowWrap, startTime uint64) *windowWrap {
	ww.windowStart = startTime
	ww.value.Store(new(int64))
	return ww
}

func Test_leapArray_calculateTimeIdx_normal(t *testing.T) {
	type fields struct {
		windowLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *atomicWindowWrapArray
		mux              triableMutex
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
				windowLengthInMs: WindowLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{}),
				mux:              triableMutex{},
			},
			args: args{
				timeMillis: 1576296044907,
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &leapArray{
				windowLengthInMs: tt.fields.windowLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       tt.fields.mux,
			}
			if got := la.calculateTimeIdx(tt.args.timeMillis); got != tt.want {
				t.Errorf("leapArray.calculateTimeIdx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateStartTime_normal(t *testing.T) {
	type fields struct {
	}
	type args struct {
		timeMillis       uint64
		windowLengthInMs uint32
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
				windowLengthInMs: WindowLengthInMs,
			},
			want: 1576296044500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStartTime(tt.args.timeMillis, tt.args.windowLengthInMs); got != tt.want {
				t.Errorf("leapArray.calculateStartTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_leapArray_WindowStartCheck_normal(t *testing.T) {
	type fields struct {
		windowLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *atomicWindowWrapArray
		mux              triableMutex
	}
	type args struct {
		bg         bucketGenerator
		timeMillis uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64 //start time of window
	}{
		{
			name: "Test_leapArray_WindowStartCheck_normal",
			fields: fields{
				windowLengthInMs: WindowLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{}),
				mux:              triableMutex{},
			},
			args: args{
				bg:         new(leapArrayMock),
				timeMillis: 1576296044907,
			},
			want: 1576296044500,
		},
	}
	wwPtr := tests[0].fields.array.get(9)
	wwPtr.windowStart = 1576296044500 //start time of cycle 1576296040000

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &leapArray{
				windowLengthInMs: tt.fields.windowLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       tt.fields.mux,
			}
			got, err := la.currentWindowWithTime(tt.args.timeMillis, tt.args.bg)
			if err != nil {
				t.Errorf("leapArray.currentWindowWithTime() error = %v\n", err)
				return
			}
			if got.windowStart != tt.want {
				t.Errorf("windowStart = %v, want %v", got.windowStart, tt.want)
			}
		})
	}
}

func Test_leapArray_currentWindowWithTime_normal(t *testing.T) {
	type fields struct {
		windowLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *atomicWindowWrapArray
		mux              triableMutex
	}
	type args struct {
		bg         bucketGenerator
		timeMillis uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *windowWrap
		wantErr bool
	}{
		{
			name: "Test_leapArray_currentWindowWithTime_normal",
			fields: fields{
				windowLengthInMs: WindowLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{}),
				mux:              triableMutex{},
			},
			args: args{
				bg:         new(leapArrayMock),
				timeMillis: 1576296044907,
			},
			want:    nil,
			wantErr: false,
		},
	}

	wwPtr := tests[0].fields.array.get(9)
	wwPtr.windowStart = 1576296044500 //start time of cycle 1576296040000
	tests[0].want = tests[0].fields.array.get(9)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &leapArray{
				windowLengthInMs: tt.fields.windowLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       tt.fields.mux,
			}
			got, err := la.currentWindowWithTime(tt.args.timeMillis, tt.args.bg)
			if (err != nil) != tt.wantErr {
				t.Errorf("leapArray.currentWindowWithTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("leapArray.currentWindowWithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_leapArray_valuesWithTime_normal(t *testing.T) {
	type fields struct {
		windowLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *atomicWindowWrapArray
		mux              triableMutex
	}
	type args struct {
		timeMillis uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *windowWrap
		wantErr bool
	}{
		{
			name: "Test_leapArray_valuesWithTime_normal",
			fields: fields{
				windowLengthInMs: WindowLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{}),
				mux:              triableMutex{},
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
		ww.windowStart = start
		start += 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &leapArray{
				windowLengthInMs: tt.fields.windowLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       tt.fields.mux,
			}
			got := la.valuesWithTime(tt.args.timeMillis)
			for _, g := range got {
				find := false
				for i := 0; i < tests[0].fields.array.length; i++ {
					w := tests[0].fields.array.get(i)
					if w.windowStart == g.windowStart {
						find = true
						break
					}
				}
				if !find {
					t.Errorf("leapArray.valuesWithTime() fail")
				}
			}
		})
	}
}

func Test_leapArray_isWindowDeprecated_normal(t *testing.T) {
	type fields struct {
		windowLengthInMs uint32
		sampleCount      uint32
		intervalInMs     uint32
		array            *atomicWindowWrapArray
		mux              triableMutex
	}
	type args struct {
		startTime uint64
		ww        *windowWrap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Test_leapArray_isWindowDeprecated_normal",
			fields: fields{
				windowLengthInMs: WindowLengthInMs,
				sampleCount:      SampleCount,
				intervalInMs:     IntervalInMs,
				array:            newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{}),
				mux:              triableMutex{},
			},
			args: args{
				startTime: 1576296044907,
				ww: &windowWrap{
					windowStart: 1576296004907,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &leapArray{
				windowLengthInMs: tt.fields.windowLengthInMs,
				sampleCount:      tt.fields.sampleCount,
				intervalInMs:     tt.fields.intervalInMs,
				array:            tt.fields.array,
				updateLock:       tt.fields.mux,
			}
			if got := la.isWindowDeprecated(tt.args.startTime, tt.args.ww); got != tt.want {
				t.Errorf("leapArray.isWindowDeprecated() = %v, want %v", got, tt.want)
			}
		})
	}
}
