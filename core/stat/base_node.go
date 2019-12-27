package stat

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
	"sync/atomic"
)

type BaseStatNode struct {
	goroutineNum int32
}

func (n *BaseStatNode) MetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	panic("implement me")
}

func (n *BaseStatNode) TotalQPS() float64 {
	panic("implement me")
}

func (n *BaseStatNode) PassQPS() float64 {
	panic("implement me")
}

func (n *BaseStatNode) BlockQPS() float64 {
	panic("implement me")
}

func (n *BaseStatNode) CompleteQPS() float64 {
	panic("implement me")
}

func (n *BaseStatNode) ErrorQPS() float64 {
	panic("implement me")
}

func (n *BaseStatNode) AvgRT() float64 {
	panic("implement me")
}

func (n *BaseStatNode) MinRT() float64 {
	panic("implement me")
}

func (n *BaseStatNode) CurrentGoroutineNum() int32 {
	return atomic.LoadInt32(&(n.goroutineNum))
}

func (n *BaseStatNode) AddPassRequest(count uint64) {
	panic("implement me")
}

func (n *BaseStatNode) AddRtAndCompleteRequest(rt, count uint64) {
	panic("implement me")
}

func (n *BaseStatNode) AddBlockRequest(count uint64) {
	panic("implement me")
}

func (n *BaseStatNode) AddErrorRequest(count uint64) {
	panic("implement me")
}

func (n *BaseStatNode) IncreaseGoroutineNum() {
	atomic.AddInt32(&(n.goroutineNum), 1)
}

func (n *BaseStatNode) DecreaseGoroutineNum() {
	atomic.AddInt32(&(n.goroutineNum), -1)
}

func (n *BaseStatNode) Reset() {
	panic("implement me")
}
