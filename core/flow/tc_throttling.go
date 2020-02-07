package flow

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

// ThrottlingChecker limits the time interval between two requests.
type ThrottlingChecker struct {
	maxQueueingTimeMs uint32
	qps               float64
}

func NewThrottlingChecker(timeoutMs uint32, qps float64) *ThrottlingChecker {
	return &ThrottlingChecker{maxQueueingTimeMs: timeoutMs, qps: qps}
}

func (c *ThrottlingChecker) DoCheck(node base.StatNode, acquireCount uint32, threshold float64) *base.TokenResult {
	// TODO
	panic("implement me")
}
