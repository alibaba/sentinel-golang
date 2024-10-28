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
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
)

var (
	// resource name --->  node retryer
	retryers     = make(map[string]*Retryer)
	retryerMutex = new(sync.Mutex)
	retryerCh    = make(chan task, capacity)
)

func init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Error(fmt.Errorf("%+v", err), "Unexpected panic when consuming retryerCh")
			}
		}()
		for task := range retryerCh {
			retryer := getRetryerOfResource(task.resource)
			retryer.scheduleNodes(task.nodes)
		}
	}()
}

// Each service should have its own Retryer to proactively retry in case of node failure.
type Retryer struct {
	resource    string
	interval    time.Duration // initial value of the retry interval
	maxAttempts uint32
	counts      map[string]uint32 // ip address ---> retried count
	checkFunc   RecoveryCheckFunc
	mtx         sync.Mutex
}

func getRetryerOfResource(resource string) *Retryer {
	retryerMutex.Lock()
	defer retryerMutex.Unlock()
	if _, ok := retryers[resource]; !ok {
		retryer := &Retryer{
			resource: resource,
			counts:   make(map[string]uint32),
		}
		rule := getOutlierRuleOfResource(resource)
		if rule == nil {
			logging.Error(errors.New("nil outlier rule"), "Nil outlier rule in getRetryerOfResource()")
		} else {
			retryer.maxAttempts = rule.MaxRecoveryAttempts
			retryer.interval = time.Duration(rule.RecoveryIntervalMs * 1e6)
			if rule.RecoveryCheckFunc != nil {
				retryer.checkFunc = rule.RecoveryCheckFunc
			} else {
				retryer.checkFunc = isPortOpen
			}
		}
		retryers[resource] = retryer
	}
	return retryers[resource]
}

func isPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func (r *Retryer) scheduleNodes(nodes []string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, node := range nodes {
		if _, ok := r.counts[node]; !ok {
			r.counts[node] = 1
			logging.Info("[Outlier Retryer] Reconnecting...", "node", node)
			nodeCopy := node // Copy values to correctly capture the closure for node.
			time.AfterFunc(r.interval, func() {
				r.connectNode(nodeCopy)
			})
		}
	}
}

func (r *Retryer) connectNode(node string) {
	start := time.Now()
	if r.checkFunc(node) {
		end := time.Now()
		r.onConnected(node, uint64(end.Sub(start).Milliseconds()))
	} else {
		r.onDisconnected(node)
	}
}

func (r *Retryer) onConnected(node string, rt uint64) {
	r.mtx.Lock()
	delete(r.counts, node)
	r.mtx.Unlock()
	recycler := getRecyclerOfResource(r.resource)
	recycler.recover(node)
	breakers := getNodeBreakersOfResource(r.resource)
	if breaker, ok := breakers[node]; ok {
		breaker.OnRequestComplete(rt, nil)
	} else {
		logging.Warn("[Outlier Retryer] Failed to update status after reconnection", "node", node)
	}
}

func (r *Retryer) onDisconnected(node string) {
	r.mtx.Lock()
	r.counts[node]++
	count := r.counts[node]
	if count > r.maxAttempts {
		count = r.maxAttempts
	}
	r.mtx.Unlock()
	// Fix bugs: When multiple active checks still do not recover, it is necessary to delete node from r.counts.
	time.AfterFunc(r.interval*time.Duration(count), func() {
		r.connectNode(node)
	})
}
