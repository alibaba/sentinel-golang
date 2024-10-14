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
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newRecycler(resource string, interval time.Duration) *Recycler {
	return &Recycler{
		resource: resource,
		status:   make(map[string]bool),
		interval: interval,
	}
}

func (r *Recycler) length() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return len(r.status)
}

func addOutlierRuleForRecycler(resource string, seconds uint32) {
	updateMux.Lock()
	defer updateMux.Unlock()
	outlierRules[resource] = &Rule{RecycleIntervalS: seconds}
}

// randomSample selects a random sample of nodes without modifying the original slice
func randomSample(nodes []string, sampleSize int) []string {
	if sampleSize > len(nodes) {
		sampleSize = len(nodes)
	}
	startIndex := rand.Intn(len(nodes)) // Select a random starting index
	result := make([]string, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		result = append(result, nodes[(startIndex+i)%len(nodes)])
	}
	return result
}

func generateNodes(n int) []string {
	nodes := make([]string, n)
	for i := 0; i < n; i++ {
		nodes[i] = fmt.Sprintf("node%d", i)
	}
	return nodes
}

func testRecycler(t *testing.T) {
	nodes := []string{"node0", "node1"}
	resource := "testResource"
	addNodeBreakers(resource, nodes)

	recycler := newRecycler(resource, 4*time.Second)
	recycler.scheduleNodes(nodes)
	assert.Equal(t, len(nodes), recycler.length())

	// Restore node0 after 2 seconds of simulation
	time.AfterFunc(2*time.Second, func() {
		recycler.recover(nodes[0])
	})
	time.Sleep(5 * time.Second)

	m := getNodeBreakersOfResource(resource)
	assert.Equal(t, 1, len(m))
	assert.Contains(t, m, nodes[0]) // node0 should have been recovered
}

func testRecyclerConcurrent(t *testing.T) {
	nodes := generateNodes(100) // Generate 100 nodes
	resource := "testResource"
	addNodeBreakers(resource, nodes)

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	recycler := newRecycler(resource, 4*time.Second)

	// Start multiple goroutines to schedule random nodes
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			numToSchedule := len(nodes) / numGoroutines * 2
			selectedNodes := randomSample(nodes, numToSchedule)
			recycler.scheduleNodes(selectedNodes)
		}()
	}
	wg.Wait()

	// Check the status of nodes
	assert.GreaterOrEqual(t, len(nodes), recycler.length())
	recycler.scheduleNodes(nodes)
	assert.Equal(t, len(nodes), recycler.length())

	// Recover node0 and node1 after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		recycler.recover(nodes[0])
		recycler.recover(nodes[1])
	})
	time.Sleep(5 * time.Second)

	// node0 and node1 should be recovered
	m := getNodeBreakersOfResource(resource)
	assert.Equal(t, 2, len(m))
	assert.Contains(t, m, nodes[0])
	assert.Contains(t, m, nodes[1])
}

func testRecyclerCh(t *testing.T) {
	nodes := []string{"node0", "node1"}
	resource := "testResource"
	addNodeBreakers(resource, nodes)
	addOutlierRuleForRecycler(resource, 4)

	recyclerCh <- task{nodes, resource}

	// Restore node0 after 2 seconds of simulation
	time.AfterFunc(2*time.Second, func() {
		recycler := getRecyclerOfResource(resource)
		recycler.recover(nodes[0])
	})
	time.Sleep(5 * time.Second)

	m := getNodeBreakersOfResource(resource)
	assert.Equal(t, 1, len(m))
	assert.Contains(t, m, nodes[0]) // node0 should have been recovered
}

func testRecyclerChConcurrent(t *testing.T) {
	nodes := generateNodes(100) // Generate 100 nodes
	resource := "testResource"
	addNodeBreakers(resource, nodes)
	addOutlierRuleForRecycler(resource, 4)

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	recycler := getRecyclerOfResource(resource)
	// Start multiple goroutines to schedule random nodes
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			numToSchedule := len(nodes) / numGoroutines * 2
			selectedNodes := randomSample(nodes, numToSchedule)
			recyclerCh <- task{selectedNodes, resource}
		}()
	}
	wg.Wait()

	// Check the status of nodes
	assert.GreaterOrEqual(t, len(nodes), recycler.length())
	recycler.scheduleNodes(nodes)
	assert.Equal(t, len(nodes), recycler.length())

	// Recover node0 and node1 after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		recycler.recover(nodes[0])
		recycler.recover(nodes[1])
	})
	time.Sleep(5 * time.Second)

	// node0 and node1 should be recovered
	m := getNodeBreakersOfResource(resource)
	assert.Equal(t, 2, len(m))
	assert.Contains(t, m, nodes[0])
	assert.Contains(t, m, nodes[1])
}

func TestRecyclerAll(t *testing.T) {
	t.Run("TestRecycler", testRecycler)
	t.Run("TestRecyclerConcurrent", testRecyclerConcurrent)
	t.Run("TestRecyclerCh", testRecyclerCh)
	t.Run("TestRecyclerChConcurrent", testRecyclerChConcurrent)
}
