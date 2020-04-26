package base

import (
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/util"
)

func Test_newAtomicBucketWrapArray_normal(t *testing.T) {
	type args struct {
		len              int
		bucketLengthInMs uint32
		bg               bucketGenerator
	}
	tests := []struct {
		name string
		args args
		want *atomicBucketWrapArray
	}{
		{
			name: "Test_newAtomicBucketWrapArray_normal",
			args: args{
				len:              int(SampleCount),
				bucketLengthInMs: BucketLengthInMs,
				bg:               &leapArrayMock{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := newAtomicBucketWrapArray(tt.args.len, tt.args.bucketLengthInMs, tt.args.bg)
			if ret == nil || uintptr(ret.base) == uintptr(0) || ret.length != tt.args.len || ret.data == nil || len(ret.data) == 0 {
				t.Errorf("newAtomicBucketWrapArray() %+v is illegal.\n", ret)
				return
			}
			dataNil := false
			for _, v := range ret.data {
				if v == nil {
					dataNil = true
					break
				}
			}
			if dataNil {
				t.Error("newAtomicBucketWrapArray exists nil bucketWrap.")
			}

		})
	}
}

func Test_atomicBucketWrapArray_elementOffset(t *testing.T) {
	type args struct {
		len              int
		bucketLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want uintptr
	}{
		{
			name: "Test_atomicBucketWrapArray_elementOffset",
			args: args{
				len:              int(SampleCount),
				bucketLengthInMs: BucketLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: 72,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicBucketWrapArray(tt.args.len, tt.args.bucketLengthInMs, tt.args.bg)
			if got := uintptr(aa.elementOffset(tt.args.idx)) - uintptr(aa.base); got != tt.want {
				t.Errorf("atomicBucketWrapArray.elementOffset() = %v, want %v \n", got, tt.want)
			}
		})
	}
}

func Test_atomicBucketWrapArray_get(t *testing.T) {
	type args struct {
		len              int
		bucketLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want *bucketWrap
	}{
		{
			name: "Test_atomicBucketWrapArray_get",
			args: args{
				len:              int(SampleCount),
				bucketLengthInMs: BucketLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicBucketWrapArray(tt.args.len, tt.args.bucketLengthInMs, tt.args.bg)
			tt.want = aa.data[9]
			if got := aa.get(tt.args.idx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("atomicBucketWrapArray.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_atomicBucketWrapArray_compareAndSet(t *testing.T) {
	type args struct {
		len              int
		bucketLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_atomicBucketWrapArray_compareAndSet",
			args: args{
				len:              int(SampleCount),
				bucketLengthInMs: BucketLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicBucketWrapArray(tt.args.len, tt.args.bucketLengthInMs, tt.args.bg)
			update := &bucketWrap{
				bucketStart: 8888888888888,
				value:       atomic.Value{},
			}
			update.value.Store(int64(666666))
			except := aa.get(9)
			if got := aa.compareAndSet(tt.args.idx, except, update); got != tt.want {
				t.Errorf("atomicBucketWrapArray.compareAndSet() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(aa.get(9), update) {
				t.Errorf("atomicBucketWrapArray.compareAndSet() update fail")
			}
		})
	}
}

func taskGet(wg *sync.WaitGroup, at *atomicBucketWrapArray, t *testing.T) {
	time.Sleep(time.Millisecond * 3)
	idx := rand.Int() % 20
	wwPtr := at.get(idx)
	vInterface := wwPtr.value.Load()
	vp, ok := vInterface.(*int64)
	if !ok {
		t.Error("bucketWrap value assert fail.\n")
	}
	v := atomic.LoadInt64(vp)
	newV := v + 1
	for !atomic.CompareAndSwapInt64(vp, v, newV) {
		v = atomic.LoadInt64(vp)
		newV = v + 1
	}
	wg.Done()
}

func Test_atomicBucketWrapArray_Concurrency_Get(t *testing.T) {
	ret := newAtomicBucketWrapArray(int(SampleCount), BucketLengthInMs, &leapArrayMock{})
	for _, ww := range ret.data {
		c := new(int64)
		*c = 0
		ww.value.Store(c)
	}
	const GoroutineNum = 1000
	wg1 := &sync.WaitGroup{}
	wg1.Add(GoroutineNum)
	for i := 0; i < GoroutineNum; i++ {
		go taskGet(wg1, ret, t)
	}
	wg1.Wait()
	sum := int64(0)
	for _, ww := range ret.data {
		val := ww.value.Load()
		count, ok := val.(*int64)
		if !ok {
			t.Error("assert error")
		}
		sum += *count
	}
	if sum != GoroutineNum {
		t.Error("sum error")
	}
	t.Log("all done")
}

func taskSet(wg *sync.WaitGroup, at *atomicBucketWrapArray, t *testing.T) {
	time.Sleep(time.Millisecond * 3)
	idx := rand.Int() % 20
	ww := at.get(idx)
	bucket := new(int64)
	*bucket = 100
	val := atomic.Value{}
	val.Store(bucket)
	replace := &bucketWrap{
		bucketStart: util.CurrentTimeMillis(),
		value:       val,
	}
	for !at.compareAndSet(idx, ww, replace) {
		ww = at.get(idx)
	}
	wg.Done()
}

func Test_atomicBucketWrapArray_Concurrency_Set(t *testing.T) {
	ret := newAtomicBucketWrapArray(int(SampleCount), BucketLengthInMs, &leapArrayMock{})
	for _, ww := range ret.data {
		c := new(int64)
		*c = 0
		ww.value.Store(c)
	}
	const GoroutineNum = 1000
	wg2 := &sync.WaitGroup{}
	wg2.Add(GoroutineNum)

	for i := 0; i < GoroutineNum; i++ {
		go taskSet(wg2, ret, t)
	}
	wg2.Wait()
	for _, ww := range ret.data {
		v := ww.value.Load()
		val, ok := v.(*int64)
		if !ok || *val != 100 {
			t.Error("assert error")
		}
	}
	t.Log("all done")
}
