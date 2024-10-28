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
	"sync"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
)

const capacity = 200

var (
	// resource name --->  node recycler
	recyclers     = make(map[string]*Recycler)
	recyclerMutex = new(sync.Mutex)
	recyclerCh    = make(chan task, capacity)
)

type task struct {
	nodes    []string
	resource string
}

func init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Error(fmt.Errorf("%+v", err), "Unexpected panic when consuming recyclerCh")
			}
		}()
		for task := range recyclerCh {
			recycler := getRecyclerOfResource(task.resource)
			recycler.scheduleNodes(task.nodes)
		}
	}()
}

// Recycler recycles node instance that have been invalidated for a long time
type Recycler struct {
	resource string
	interval time.Duration
	status   map[string]bool
	mtx      sync.Mutex
}

func getRecyclerOfResource(resource string) *Recycler {
	recyclerMutex.Lock()
	defer recyclerMutex.Unlock()
	if _, ok := recyclers[resource]; !ok {
		recycler := &Recycler{
			resource: resource,
			status:   make(map[string]bool),
		}
		rule := getOutlierRuleOfResource(resource)
		if rule == nil {
			logging.Error(errors.New("nil outlier rule"), "Nil outlier rule in getRecyclerOfResource()")
		} else {
			if rule.RecycleIntervalS == 0 {
				recycler.interval = 10 * time.Minute
			} else {
				recycler.interval = time.Duration(rule.RecycleIntervalS * 1e9)
			}
		}
		recyclers[resource] = recycler
	}
	return recyclers[resource]
}

func (r *Recycler) scheduleNodes(nodes []string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, node := range nodes {
		if _, ok := r.status[node]; !ok {
			r.status[node] = false
			nodeCopy := node // Copy values to correctly capture the closure for node.
			time.AfterFunc(r.interval, func() {
				r.recycle(nodeCopy)
			})
		}
	}
}

func (r *Recycler) recover(node string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if _, ok := r.status[node]; ok {
		r.status[node] = true
	}
}

func (r *Recycler) recycle(node string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if v, ok := r.status[node]; ok && !v {
		deleteNodeBreakerOfResource(r.resource, node)
	}
	delete(r.status, node)
}
