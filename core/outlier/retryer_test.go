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
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

type dummyCall struct {
	callCounts   map[string]int
	recoverCount int
	mtx          sync.Mutex
}

func newDummyCall() *dummyCall {
	return &dummyCall{
		callCounts: make(map[string]int),
	}
}

func (d *dummyCall) registerAddress(address string, n int) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.callCounts[address] = n
}

// dummyCall's Check checks whether the node address has returned to normal.
// It returns to normal when the value recorded in callCounts decreases to 0.
func (d *dummyCall) Check(address string) bool {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if _, ok := d.callCounts[address]; ok {
		d.callCounts[address]--
		time.Sleep(100 * time.Millisecond) // simulate network latency
		if d.callCounts[address] == 0 {
			fmt.Printf("%s successfully reconnected\n", address)
			d.recoverCount++
			return true
		}
		return false
	}
	panic("Attempting to call an unregistered node address.")
}

func (d *dummyCall) getRecoverCount() int {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	return d.recoverCount
}

func addOutlierRuleForRetryer(resource string, n, internal uint32, f RecoveryCheckFunc) {
	updateMux.Lock()
	defer updateMux.Unlock()
	outlierRules[resource] = &Rule{
		MaxRecoveryAttempts: n,
		RecoveryIntervalMs:  internal,
		RecoveryCheckFunc:   f,
	}
}

// MockCircuitBreaker is a mock implementation of CircuitBreaker
type MockCircuitBreaker struct {
	mock.Mock
}

func (m *MockCircuitBreaker) BoundRule() *circuitbreaker.Rule {
	args := m.Called()
	return args.Get(0).(*circuitbreaker.Rule)
}

func (m *MockCircuitBreaker) BoundStat() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockCircuitBreaker) TryPass(ctx *base.EntryContext) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockCircuitBreaker) CurrentState() circuitbreaker.State {
	args := m.Called()
	return args.Get(0).(circuitbreaker.State)
}

func (m *MockCircuitBreaker) OnRequestComplete(rtt uint64, err error) {
	m.Called(rtt, err)
}

func addNodeBreakers(resource string, nodes []string) {
	updateMux.Lock()
	defer updateMux.Unlock()
	if nodeBreakers[resource] == nil {
		nodeBreakers[resource] = make(map[string]circuitbreaker.CircuitBreaker)
	}
	for _, address := range nodes {
		nodeBreakers[resource][address] = &MockCircuitBreaker{}
	}
}

func setNodeBreaker(resource string, node string, breaker *MockCircuitBreaker) {
	updateMux.Lock()
	defer updateMux.Unlock()
	if nodeBreakers[resource] == nil {
		nodeBreakers[resource] = make(map[string]circuitbreaker.CircuitBreaker)
	}
	nodeBreakers[resource][node] = breaker
}

func (r *Retryer) length() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return len(r.counts)
}

// Construct two dummy node addresses: the first one recovers after the third check,
// and the second one recovers after math.MaxInt32 checks. Observe the changes in the
// circuit breaker and callCounts status for the first node before and after recovery.
func testRetryer(t *testing.T) {
	resource := "testResource0"
	nodes := []string{"node0", "node1"}
	var internal, n uint32 = 1000, 3
	d := newDummyCall()
	d.registerAddress(nodes[0], int(n))
	d.registerAddress(nodes[1], math.MaxInt32)

	addOutlierRuleForRetryer(resource, n, internal, d.Check)
	retryer := getRetryerOfResource(resource)
	retryer.scheduleNodes(nodes)

	mockCB := new(MockCircuitBreaker)
	mockCB.On("OnRequestComplete", mock.AnythingOfType("uint64"), nil).Return()
	setNodeBreaker(resource, nodes[0], mockCB)

	minDuration := time.Duration(n * (n + 1) / 2 * internal * 1e6)
	for d.getRecoverCount() < 1 {
		time.Sleep(minDuration)
	}
	assert.Equal(t, len(nodes)-1, retryer.length())
	mockCB.AssertExpectations(t)
}

func testRetryerConcurrent(t *testing.T) {
	resource := "testResource1"
	nodes := generateNodes(100) // Generate 100 nodes
	var internal, n uint32 = 1000, 3
	d := newDummyCall()
	mockCBs := make([]*MockCircuitBreaker, 0, len(nodes)/2)
	for i, node := range nodes {
		if i%2 == 0 {
			mockCB := new(MockCircuitBreaker)
			mockCB.On("OnRequestComplete", mock.AnythingOfType("uint64"), nil).Return()
			setNodeBreaker(resource, node, mockCB)
			mockCBs = append(mockCBs, mockCB)
			d.registerAddress(node, int(n))
		} else {
			d.registerAddress(node, math.MaxInt32)
		}
	}

	addOutlierRuleForRetryer(resource, n, internal, d.Check)
	retryer := getRetryerOfResource(resource)
	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines to schedule random nodes
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			numToSchedule := len(nodes) / numGoroutines * 2
			selectedNodes := randomSample(nodes, numToSchedule)
			retryer.scheduleNodes(selectedNodes)
		}()
	}
	wg.Wait()

	// Check the status of nodes
	assert.GreaterOrEqual(t, len(nodes), retryer.length())
	retryer.scheduleNodes(nodes)
	assert.Equal(t, len(nodes), retryer.length())

	minDuration := time.Duration(n * (n + 1) / 2 * internal * 1e6)
	for d.getRecoverCount() < len(nodes)/2 {
		time.Sleep(minDuration)
	}
	assert.Equal(t, len(nodes)/2, retryer.length())
	for _, breaker := range mockCBs {
		breaker.AssertExpectations(t)
	}
}

func testRetryerCh(t *testing.T) {
	nodes := []string{"node0", "node1"}
	resource := "testResource2"
	var internal, n uint32 = 1000, 3
	d := newDummyCall()
	d.registerAddress(nodes[0], int(n))
	d.registerAddress(nodes[1], math.MaxInt32)

	addOutlierRuleForRetryer(resource, n, internal, d.Check)
	retryer := getRetryerOfResource(resource)

	mockCB := new(MockCircuitBreaker)
	mockCB.On("OnRequestComplete", mock.AnythingOfType("uint64"), nil).Return()
	setNodeBreaker(resource, nodes[0], mockCB)

	retryerCh <- task{nodes, resource}

	minDuration := time.Duration(n * (n + 1) / 2 * internal * 1e6)
	for d.getRecoverCount() < 1 {
		time.Sleep(minDuration)
	}
	assert.Equal(t, len(nodes)-1, retryer.length())
	mockCB.AssertExpectations(t)
}

func TestRetryerAll(t *testing.T) {
	t.Run("TestRetryer", testRetryer)
	t.Run("TestRetryerConcurrent", testRetryerConcurrent)
	t.Run("TestRetryerCh", testRetryerCh)
}
