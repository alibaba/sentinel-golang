// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outlier

import (
	"net"
	"sync"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
)

// resource name --->  node retryer
var retryers = make(map[string]*Retryer)
var updateMutex = new(sync.Mutex)

// Each service should have its own Retryer to proactively
// retry in case of node failure.
type Retryer struct {
	resource    string
	interval    time.Duration
	maxAttempts uint32
	// TODO RWMutex counts
	counts map[string]uint32 // ip address ---> retry count
}

func getRetryerOfResource(resource string) *Retryer {
	updateMutex.Lock()
	defer updateMutex.Unlock()
	if _, ok := retryers[resource]; !ok {
		retryer := &Retryer{resource: resource}
		rules := getOutlierRuleOfResource(resource)
		if rules != nil {
			retryer.maxAttempts = rules.MaxRecoveryAttempts // TODO per resource only has one rule
			retryer.interval = time.Duration(rules.RecoveryInterval * 1e6)
			retryer.counts = make(map[string]uint32)
		}
		retryers[resource] = retryer
	}
	return retryers[resource]
}

func (r *Retryer) ConnectNode(nodeID string) {
	ok, rt := isPortOpen(nodeID)
	if ok {
		delete(r.counts, nodeID)
		r.OnCompleted(nodeID, rt)
	} else {
		r.counts[nodeID]++
		count := r.counts[nodeID]
		if count > r.maxAttempts {
			count = r.maxAttempts
		}
		time.AfterFunc(r.interval*time.Duration(count), func() {
			r.ConnectNode(nodeID)
		})
	}
}

func (r *Retryer) scheduleRetry(nodes []string) {
	for _, node := range nodes {
		if _, ok := r.counts[node]; !ok {
			logging.Debug("[Outlier Reconnect]", "nodeID", node)
			r.counts[node] = 1
			time.AfterFunc(r.interval, func() {
				r.ConnectNode(node)
			})
		}
	}
}

// TODO WithCheckFunc
func isPortOpen(address string) (bool, uint64) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false, 0
	}
	defer conn.Close()
	end := time.Now()
	return true, uint64(end.Sub(start).Milliseconds())
}

func (r *Retryer) OnCompleted(nodeID string, rt uint64) {
	recyclers[r.resource].recover(nodeID)
	nodes := getNodeBreakerOfResource(r.resource)
	// TODO 判断nodes[nodeID]不存在
	nodes[nodeID].OnRequestComplete(rt, nil)
}
