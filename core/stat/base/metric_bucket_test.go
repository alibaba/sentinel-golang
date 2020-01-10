package base

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"sync"
	"testing"
	"unsafe"
)

func Test_metricBucket_MemSize(t *testing.T) {
	mb := newMetricBucket()
	size := unsafe.Sizeof(*mb)
	if size != 48 {
		t.Error("unexpect memory size of metricBucket")
	}
}

func Test_metricBucket_Normal(t *testing.T) {
	mb := newMetricBucket()

	for i := 0; i < 100; i++ {
		if i%5 == 0 {
			mb.addPass(1)
		} else if i%5 == 1 {
			mb.addBlock(1)
		} else if i%5 == 2 {
			mb.addComplete(1)
		} else if i%5 == 3 {
			mb.addError(1)
		} else if i%5 == 4 {
			mb.addRt(100)
		} else {
			t.Error("unexpect idx")
		}
	}

	if mb.get(base.MetricEventPass) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.get(base.MetricEventBlock) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.get(base.MetricEventComplete) != 20 {
		t.Error("unexpect count MetricEventComplete")
	}
	if mb.get(base.MetricEventError) != 20 {
		t.Error("unexpect count MetricEventError")
	}
	if mb.get(base.MetricEventRt) != 20*100 {
		t.Error("unexpect count MetricEventRt")
	}
}

func Test_metricBucket_Concurrent(t *testing.T) {
	mb := newMetricBucket()
	wg := &sync.WaitGroup{}
	wg.Add(5000)

	for i := 0; i < 1000; i++ {
		go func() {
			mb.addPass(1)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.addBlock(2)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.addComplete(3)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.addError(4)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func(c uint64) {
			mb.addRt(int64(c))
			wg.Done()
		}(uint64(i))
	}
	wg.Wait()

	if mb.get(base.MetricEventPass) != 1000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.get(base.MetricEventBlock) != 2000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.get(base.MetricEventComplete) != 3000 {
		t.Error("unexpect count MetricEventComplete")
	}
	if mb.get(base.MetricEventError) != 4000 {
		t.Error("unexpect count MetricEventError")
	}

	totalRt := (0 + 999) * 1000 / 2
	if mb.get(base.MetricEventRt) != int64(totalRt) {
		t.Error("unexpect count MetricEventRt")
	}
}
