package base

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/alibaba/sentinel-golang/core/base"
)

func Test_metricBucket_MemSize(t *testing.T) {
	mb := NewMetricBucket()
	size := unsafe.Sizeof(*mb)
	if size != 48 {
		t.Error("unexpect memory size of MetricBucket")
	}
}

func Test_metricBucket_Normal(t *testing.T) {
	mb := NewMetricBucket()

	for i := 0; i < 100; i++ {
		if i%5 == 0 {
			mb.Add(base.MetricEventPass, 1)
		} else if i%5 == 1 {
			mb.Add(base.MetricEventBlock, 1)
		} else if i%5 == 2 {
			mb.Add(base.MetricEventComplete, 1)
		} else if i%5 == 3 {
			mb.Add(base.MetricEventError, 1)
		} else if i%5 == 4 {
			mb.AddRt(100)
		} else {
			t.Error("unexpect idx")
		}
	}

	if mb.Get(base.MetricEventPass) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(base.MetricEventBlock) != 20 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(base.MetricEventComplete) != 20 {
		t.Error("unexpect count MetricEventComplete")
	}
	if mb.Get(base.MetricEventError) != 20 {
		t.Error("unexpect count MetricEventError")
	}
	if mb.Get(base.MetricEventRt) != 20*100 {
		t.Error("unexpect count MetricEventRt")
	}
}

func Test_metricBucket_Concurrent(t *testing.T) {
	mb := NewMetricBucket()
	wg := &sync.WaitGroup{}
	wg.Add(5000)

	for i := 0; i < 1000; i++ {
		go func() {
			mb.Add(base.MetricEventPass, 1)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.Add(base.MetricEventBlock, 2)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.Add(base.MetricEventComplete, 3)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func() {
			mb.Add(base.MetricEventError, 4)
			wg.Done()
		}()
	}

	for i := 0; i < 1000; i++ {
		go func(c uint64) {
			mb.AddRt(int64(c))
			wg.Done()
		}(uint64(i))
	}
	wg.Wait()

	if mb.Get(base.MetricEventPass) != 1000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(base.MetricEventBlock) != 2000 {
		t.Error("unexpect count MetricEventBlock")
	}
	if mb.Get(base.MetricEventComplete) != 3000 {
		t.Error("unexpect count MetricEventComplete")
	}
	if mb.Get(base.MetricEventError) != 4000 {
		t.Error("unexpect count MetricEventError")
	}

	totalRt := (0 + 999) * 1000 / 2
	if mb.Get(base.MetricEventRt) != int64(totalRt) {
		t.Error("unexpect count MetricEventRt")
	}
}
