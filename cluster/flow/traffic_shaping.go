package flow

import (
	"runtime"
	"sync/atomic"

	"github.com/alibaba/sentinel-golang/cluster/client"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

const clusterFlowBlockedMsg = "cluster flow traffic shaping reject request"

type TrafficShapingController struct {
	rule               *ClusterRule
	mux                util.Mutex
	localTokenSequence int64
	tokenService       client.TokenService
}

func (tc *TrafficShapingController) BoundRule() *ClusterRule {
	return tc.rule
}

func (tc *TrafficShapingController) DoCheck(res string, batchCount uint32) *base.TokenResult {
	r := tc.rule

	// acquire tokens directly from token server
	if r.TokenSequence == 0 || r.TokenSequence == 1 {
		curCount, err := tc.tokenService.AcquireFlowToken(res, batchCount, r.StatIntervalInMs)
		if err != nil {
			logging.Error(err, "Fail to fetch token from token server in TrafficShapingController#DoCheck", "rule", r)
			return nil
		}
		if curCount > r.Threshold {
			return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, clusterFlowBlockedMsg, tc.rule, curCount)
		}
		return nil
	}

	// acquire tokens from local token sequence.
	// if local token sequence is drained, reacquire token sequence from token server
	for {
		// acquired token from local token sequence.
		if atomic.AddInt64(&tc.localTokenSequence, -int64(batchCount)) > 0 {
			return nil
		}

		if tc.mux.TryLock() {
			curGlobalCount, err := tc.tokenService.AcquireFlowToken(res, r.TokenSequence, r.StatIntervalInMs)
			// if token server is unavailable, pass checking directly
			if err != nil {
				tc.mux.Unlock()
				logging.Error(err, "Fail to reacquire token from token server in TrafficShapingController#doReacquireTokenSequence", "rule", r)
				return nil
			}
			if curGlobalCount > r.Threshold {
				tc.mux.Unlock()
				return base.NewTokenResultBlockedWithCause(base.BlockTypeFlow, clusterFlowBlockedMsg, tc.rule, curGlobalCount)
			}
			atomic.StoreInt64(&tc.localTokenSequence, int64(r.TokenSequence))
		} else {
			runtime.Gosched()
		}
	}
}
