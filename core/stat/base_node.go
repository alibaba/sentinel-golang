package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	sbase "github.com/sentinel-group/sentinel-golang/core/stat/base"
	"sync/atomic"
)

type BaseStatNode struct {
	goroutineNum   int32
	sampleCount    uint32
	intervalInMs   uint32
	rollingCounter *sbase.BucketLeapArray
}

func NewBaseStatNode(sampleCount uint32, intervalInMs uint32) *BaseStatNode {
	return &BaseStatNode{
		goroutineNum:   0,
		sampleCount:    sampleCount,
		intervalInMs:   intervalInMs,
		rollingCounter: sbase.NewBucketLeapArray(sampleCount, intervalInMs),
	}
}

func (n *BaseStatNode) MetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	panic("implement me")
}

func (n *BaseStatNode) TotalQPS() float64 {
	return n.GetQPS(base.MetricEventPass) +n.GetQPS(base.MetricEventBlock)
}

func (n *BaseStatNode) GetQPS(event base.MetricEvent) float64 {
	return float64(n.rollingCounter.Count(event)) / n.rollingCounter.GetIntervalInSecond()
}

func (n *BaseStatNode) AddRequest(event base.MetricEvent, count uint64) {
	n.rollingCounter.AddCount(event, int64(count))
}

func (n *BaseStatNode) AddRtAndCompleteRequest(rt, count uint64) {
	n.rollingCounter.AddCount(base.MetricEventComplete, int64(count))
	n.rollingCounter.AddCount(base.MetricEventRt, int64(count))
}

func (n *BaseStatNode) AvgRT() float64 {
	complete := n.rollingCounter.Count(base.MetricEventComplete)
	if complete <= 0 {
		return float64(0)
	}
	return float64(n.rollingCounter.Count(base.MetricEventRt) / complete)
}

func (n *BaseStatNode) MinRT() float64 {
	return float64(n.rollingCounter.MinRt())
}

func (n *BaseStatNode) CurrentGoroutineNum() int32 {
	return atomic.LoadInt32(&(n.goroutineNum))
}

func (n *BaseStatNode) IncreaseGoroutineNum() {
	atomic.AddInt32(&(n.goroutineNum), 1)
}

func (n *BaseStatNode) DecreaseGoroutineNum() {
	atomic.AddInt32(&(n.goroutineNum), -1)
}

func (n *BaseStatNode) Reset() {
	n.rollingCounter = sbase.NewBucketLeapArray(n.sampleCount, n.intervalInMs)
}
