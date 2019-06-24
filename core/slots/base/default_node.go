package base

import (
	"github.com/sentinel-group/sentinel-golang/core/node"
	"github.com/sentinel-group/sentinel-golang/core/slots/statistic/data"
	"sync/atomic"
	"time"
)

const (
	windowLengthImMs_ uint32 = 200
	sampleCount_      uint32 = 5
	intervalInMs_     uint32 = 1000
)

type DefaultNode struct {
	rollingCounterInSecond *data.SlidingWindow
	rollingCounterInMinute *data.SlidingWindow
	currentGoroutineNum    uint32
	lastFetchTime          uint64
	resourceWrapper        *ResourceWrapper
}

func NewDefaultNode(wrapper *ResourceWrapper) *DefaultNode {
	return &DefaultNode{
		rollingCounterInSecond: data.NewSlidingWindow(sampleCount_, intervalInMs_),
		rollingCounterInMinute: data.NewSlidingWindow(sampleCount_, intervalInMs_),
		currentGoroutineNum:    0,
		lastFetchTime:          uint64(time.Now().UnixNano() / (1e6)),
		resourceWrapper:        wrapper,
	}
}

func (dn *DefaultNode) AddPass(count uint64) {
	dn.rollingCounterInSecond.AddCount(data.MetricEventPass, count)
}

func (dn *DefaultNode) AddGoroutineNum(count uint32) {
	atomic.AddUint32(&dn.currentGoroutineNum, count)
}

func (dn *DefaultNode) TotalRequest() uint64 {
	return dn.rollingCounterInSecond.Count(data.MetricEventPass) + dn.rollingCounterInSecond.Count(data.MetricEventBlock)
}
func (dn *DefaultNode) TotalPass() uint64 {
	return dn.rollingCounterInMinute.Count(data.MetricEventPass)
}
func (dn *DefaultNode) TotalSuccess() uint64 {
	return dn.rollingCounterInMinute.Count(data.MetricEventSuccess)
}
func (dn *DefaultNode) BlockRequest() uint64 {
	return dn.rollingCounterInMinute.Count(data.MetricEventBlock)
}
func (dn *DefaultNode) TotalError() uint64 {
	return dn.rollingCounterInMinute.Count(data.MetricEventError)
}
func (dn *DefaultNode) PassQps() uint64 {
	return dn.rollingCounterInSecond.Count(data.MetricEventPass) / uint64(intervalInMs_)
}
func (dn *DefaultNode) BlockQps() uint64 {
	return dn.rollingCounterInSecond.Count(data.MetricEventBlock) / uint64(intervalInMs_)
}
func (dn *DefaultNode) TotalQps() uint64 {
	return dn.PassQps() + dn.BlockQps()
}
func (dn *DefaultNode) SuccessQps() uint64 {
	return dn.rollingCounterInSecond.Count(data.MetricEventSuccess) / uint64(intervalInMs_)
}
func (dn *DefaultNode) MaxSuccessQps() uint64 {
	return dn.rollingCounterInSecond.MaxSuccess() * uint64(sampleCount_)
}
func (dn *DefaultNode) ErrorQps() uint64 {
	return dn.rollingCounterInSecond.Count(data.MetricEventError) / uint64(intervalInMs_)
}

func (dn *DefaultNode) AvgRt() float32 {
	return 0
}
func (dn *DefaultNode) MinRt() float32 {
	return 0
}
func (dn *DefaultNode) CurGoroutineNum() uint64 {
	return 0
}
func (dn *DefaultNode) PreviousBlockQps() uint64 {
	return 0
}
func (dn *DefaultNode) PreviousPassQps() uint64 {
	return 0
}
func (dn *DefaultNode) Metrics() map[uint64]*node.MetricNode {
	return nil
}
