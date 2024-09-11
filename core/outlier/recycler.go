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
	"sync"
	"time"
)

// resource name --->  node recycler
var recyclers = make(map[string]*Recycler)
var recyclerMutex = new(sync.Mutex)

type Recycler struct {
	resource string
	interval time.Duration
	status   map[string]bool
}

func getRecyclerOfResource(resource string) *Recycler {
	recyclerMutex.Lock()
	defer recyclerMutex.Unlock()
	if _, ok := recyclers[resource]; !ok {
		recycler := &Recycler{resource: resource}
		rules := getOutlierRuleOfResource(resource)
		if rules != nil {
			recycler.interval = time.Duration(rules.RecoveryInterval * 1e6)
			recycler.status = make(map[string]bool)
		}
		recyclers[resource] = recycler
	}
	return recyclers[resource]
}

// The default policy is to recycle the breaker instance if the node does not recover in one hour
func (r *Recycler) scheduleRecycler(nodes []string) {
	for _, node := range nodes {
		if _, ok := r.status[node]; !ok {
			time.AfterFunc(r.interval, func() {
				r.recycle(node)
			})
		}
	}
}

func (r *Recycler) recover(node string) {
	r.status[node] = true
}

func (r *Recycler) recycle(node string) {
	if v, ok := r.status[node]; ok && !v {
		deleteNodeBreakerFromResource(r.resource, node)
		delete(r.status, node)
	}
}
