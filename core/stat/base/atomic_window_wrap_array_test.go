package base

import (
	"github.com/sentinel-group/sentinel-golang/util"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_newAtomicWindowWrapArray_normal(t *testing.T) {
	type args struct {
		len              int
		windowLengthInMs uint32
		bg               bucketGenerator
	}
	tests := []struct {
		name string
		args args
		want *atomicWindowWrapArray
	}{
		{
			name: "Test_newAtomicWindowWrapArray_normal",
			args: args{
				len:              int(SampleCount),
				windowLengthInMs: WindowLengthInMs,
				bg:               &leapArrayMock{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := newAtomicWindowWrapArray(tt.args.len, tt.args.windowLengthInMs, tt.args.bg)
			if ret == nil || uintptr(ret.base) == uintptr(0) || ret.length != tt.args.len || ret.data == nil || len(ret.data) == 0 {
				t.Errorf("newAtomicWindowWrapArray() %+v is illegal.\n", ret)
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
				t.Error("newAtomicWindowWrapArray exists nil windowWrap.")
			}

		})
	}
}

func Test_atomicWindowWrapArray_elementOffset(t *testing.T) {
	type args struct {
		len              int
		windowLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want uintptr
	}{
		{
			name: "Test_atomicWindowWrapArray_elementOffset",
			args: args{
				len:              int(SampleCount),
				windowLengthInMs: WindowLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: 72,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicWindowWrapArray(tt.args.len, tt.args.windowLengthInMs, tt.args.bg)
			if got := uintptr(aa.elementOffset(tt.args.idx)) - uintptr(aa.base); got != tt.want {
				t.Errorf("atomicWindowWrapArray.elementOffset() = %v, want %v \n", got, tt.want)
			}
		})
	}
}

func Test_atomicWindowWrapArray_get(t *testing.T) {
	type args struct {
		len              int
		windowLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want *windowWrap
	}{
		{
			name: "Test_atomicWindowWrapArray_get",
			args: args{
				len:              int(SampleCount),
				windowLengthInMs: WindowLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicWindowWrapArray(tt.args.len, tt.args.windowLengthInMs, tt.args.bg)
			tt.want = aa.data[9]
			if got := aa.get(tt.args.idx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("atomicWindowWrapArray.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_atomicWindowWrapArray_compareAndSet(t *testing.T) {
	type args struct {
		len              int
		windowLengthInMs uint32
		bg               bucketGenerator
		idx              int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_atomicWindowWrapArray_compareAndSet",
			args: args{
				len:              int(SampleCount),
				windowLengthInMs: WindowLengthInMs,
				bg:               &leapArrayMock{},
				idx:              9,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := newAtomicWindowWrapArray(tt.args.len, tt.args.windowLengthInMs, tt.args.bg)
			update := &windowWrap{
				bucketStart: 8888888888888,
				value:       atomic.Value{},
			}
			update.value.Store(int64(666666))
			except := aa.get(9)
			if got := aa.compareAndSet(tt.args.idx, except, update); got != tt.want {
				t.Errorf("atomicWindowWrapArray.compareAndSet() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(aa.get(9), update) {
				t.Errorf("atomicWindowWrapArray.compareAndSet() update fail")
			}
		})
	}
}

func taskGet(wg *sync.WaitGroup, at *atomicWindowWrapArray, t *testing.T) {
	time.Sleep(time.Millisecond * 3)
	idx := rand.Int() % 20
	wwPtr := at.get(idx)
	vInterface := wwPtr.value.Load()
	vp, ok := vInterface.(*int64)
	if !ok {
		t.Error("windowWrap value assert fail.\n")
	}
	v := atomic.LoadInt64(vp)
	newV := v + 1
	for !atomic.CompareAndSwapInt64(vp, v, newV) {
		v = atomic.LoadInt64(vp)
		newV = v + 1
	}
	wg.Done()
}

func Test_atomicWindowWrapArray_Concurrency_Get(t *testing.T) {
	ret := newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{})
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

func taskSet(wg *sync.WaitGroup, at *atomicWindowWrapArray, t *testing.T) {
	time.Sleep(time.Millisecond * 3)
	idx := rand.Int() % 20
	ww := at.get(idx)
	bucket := new(int64)
	*bucket = 100
	val := atomic.Value{}
	val.Store(bucket)
	replace := &windowWrap{
		bucketStart: util.CurrentTimeMillis(),
		value:       val,
	}
	for !at.compareAndSet(idx, ww, replace) {
		ww = at.get(idx)
	}
	wg.Done()
}

func Test_atomicWindowWrapArray_Concurrency_Set(t *testing.T) {
	ret := newAtomicWindowWrapArray(int(SampleCount), WindowLengthInMs, &leapArrayMock{})
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
