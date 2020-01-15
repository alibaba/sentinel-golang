package base

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/util"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	PtrSize = int(8)
)

// windowWrap represent a slot to record metrics
// In order to reduce the usage of memory, windowWrap don't hold length of windowWrap
// The length of windowWrap could be seen in leapArray.
// The scope of time is [startTime, startTime+windowLength)
// The size of windowWrap is 24(8+16) bytes
type windowWrap struct {
	// Start time of this windowWrap
	windowStart uint64
	// Value is the actual data structure to record the metrics.
	// Such as metricBucket, the size is 16 bytes.
	value atomic.Value
}

func (ww *windowWrap) resetTo(startTime uint64) {
	ww.windowStart = startTime
}

func (ww *windowWrap) isTimeInWindow(now uint64, windowLengthInMs uint32) bool {
	return ww.windowStart <= now && now < ww.windowStart+uint64(windowLengthInMs)
}

func calculateStartTime(now uint64, windowLengthInMs uint32) uint64 {
	return now - (now % uint64(windowLengthInMs))
}

// atomic windowWrap array to resolve race condition
// atomicWindowWrapArray can not append or delete element after initializing
type atomicWindowWrapArray struct {
	// The base address for real data array
	base unsafe.Pointer
	// The length of slice(array), it can not be modified.
	length int
	data   []*windowWrap
}

// New atomicWindowWrapArray with initializing field data
// Default, automatically initialize each windowWrap
// len: length of array
// windowLengthInMs: window length of windowWrap
// generator: generator to generate bucket
func newAtomicWindowWrapArray(len int, windowLengthInMs uint32, generator bucketGenerator) *atomicWindowWrapArray {
	ret := &atomicWindowWrapArray{
		length: len,
		data:   make([]*windowWrap, len),
	}

	// automatically initialize each windowWrap
	// tail windowWrap of data is initialized with current time
	startTime := calculateStartTime(util.CurrentTimeMillis(), windowLengthInMs)
	for i := len - 1; i >= 0; i-- {
		ww := &windowWrap{
			windowStart: startTime,
			value:       atomic.Value{},
		}
		ww.value.Store(generator.newEmptyBucket())
		ret.data[i] = ww
		startTime -= uint64(windowLengthInMs)
	}

	// calculate base address for real data array
	sliHeader := (*util.SliceHeader)(unsafe.Pointer(&ret.data))
	ret.base = unsafe.Pointer((**windowWrap)(unsafe.Pointer(sliHeader.Data)))
	return ret
}

func (aa *atomicWindowWrapArray) elementOffset(idx int) unsafe.Pointer {
	if idx >= aa.length && idx < 0 {
		panic(fmt.Sprintf("The index (%d) is out of bounds, length is %d.", idx, aa.length))
	}
	basePtr := aa.base
	return unsafe.Pointer(uintptr(basePtr) + uintptr(idx*PtrSize))
}

func (aa *atomicWindowWrapArray) get(idx int) *windowWrap {
	// aa.elementOffset(idx) return the secondary pointer of windowWrap, which is the pointer to the aa.data[idx]
	// then convert to (*unsafe.Pointer)
	return (*windowWrap)(atomic.LoadPointer((*unsafe.Pointer)(aa.elementOffset(idx))))
}

func (aa *atomicWindowWrapArray) compareAndSet(idx int, except, update *windowWrap) bool {
	// aa.elementOffset(idx) return the secondary pointer of windowWrap, which is the pointer to the aa.data[idx]
	// then convert to (*unsafe.Pointer)
	// update secondary pointer
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(aa.elementOffset(idx)), unsafe.Pointer(except), unsafe.Pointer(update))
}

// The windowWrap leap array,
// sampleCount represent the number of windowWrap
// intervalInMs represent the interval of leapArray.
// For example, windowLengthInMs is 500ms, intervalInMs is 1min, so sampleCount is 120.
type leapArray struct {
	windowLengthInMs uint32
	sampleCount      uint32
	intervalInMs     uint32
	array            *atomicWindowWrapArray
	// update lock
	updateLock triableMutex
}

func (la *leapArray) currentWindow(bg bucketGenerator) (*windowWrap, error) {
	return la.currentWindowWithTime(util.CurrentTimeMillis(), bg)
}

func (la *leapArray) currentWindowWithTime(now uint64, bg bucketGenerator) (*windowWrap, error) {
	if now < 0 {
		return nil, errors.New("Current time is less than 0.")
	}

	idx := la.calculateTimeIdx(now)
	windowStart := calculateStartTime(now, la.windowLengthInMs)

	for { //spin to get the current windowWrap
		old := la.array.get(idx)
		if old == nil {
			// because la.array.data had initiated when new la.array
			// theoretically, here is not reachable
			newWrap := &windowWrap{
				windowStart: windowStart,
				value:       atomic.Value{},
			}
			newWrap.value.Store(bg.newEmptyBucket())
			if la.array.compareAndSet(idx, nil, newWrap) {
				return newWrap, nil
			} else {
				runtime.Gosched()
			}
		} else if windowStart == atomic.LoadUint64(&old.windowStart) {
			return old, nil
		} else if windowStart > atomic.LoadUint64(&old.windowStart) {
			// current time has been next cycle of leapArray and leapArray dont't count in last cycle.
			// reset windowWrap
			if la.updateLock.TryLock() {
				old = bg.resetWindowTo(old, windowStart)
				la.updateLock.Unlock()
				return old, nil
			} else {
				runtime.Gosched()
			}
		} else if windowStart < old.windowStart {
			// used for some special case(e.g. when occupying "future" buckets).
			return nil, errors.New(fmt.Sprintf("Provided time timeMillis=%d is already behind old.windowStart=%d.", windowStart, old.windowStart))
		}
	}
}

func (la *leapArray) calculateTimeIdx(now uint64) int {
	timeId := now / uint64(la.windowLengthInMs)
	return int(timeId) % la.array.length
}

//  Get all windowWrap between [current time -1000ms, current time]
func (la *leapArray) values() []*windowWrap {
	return la.valuesWithTime(util.CurrentTimeMillis())
}

func (la *leapArray) valuesWithTime(now uint64) []*windowWrap {
	if now <= 0 {
		return make([]*windowWrap, 0)
	}
	ret := make([]*windowWrap, 0)
	for i := 0; i < la.array.length; i++ {
		ww := la.array.get(i)
		if ww == nil || la.isWindowDeprecated(now, ww) {
			continue
		}
		ret = append(ret, ww)
	}
	return ret
}

func (la *leapArray) ValuesWithConditional(now uint64, predicate base.TimePredicate) []*windowWrap {
	if now <= 0 {
		return make([]*windowWrap, 0)
	}
	ret := make([]*windowWrap, 0)
	for i := 0; i < la.array.length; i++ {
		ww := la.array.get(i)
		if ww == nil || la.isWindowDeprecated(now, ww) || !predicate(ww.windowStart) {
			continue
		}
		ret = append(ret, ww)
	}
	return ret

}

// Judge whether the windowWrap is expired
func (la *leapArray) isWindowDeprecated(now uint64, ww *windowWrap) bool {
	return (now - ww.windowStart) > uint64(la.intervalInMs)
}

// Generic interface to generate bucket
type bucketGenerator interface {
	// called when timestamp entry a new slot interval
	newEmptyBucket() interface{}

	// reset the windowWrap, clear all data of windowWrap
	resetWindowTo(ww *windowWrap, startTime uint64) *windowWrap
}
