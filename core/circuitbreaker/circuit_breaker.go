package circuitbreaker

import (
	"sync/atomic"
	"unsafe"

	"github.com/alibaba/sentinel-golang/core/base"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"
	"github.com/alibaba/sentinel-golang/util"
)

/**
  Circuit Breaker State Machine:

                                 switch to open based on rule
         +-----------------------------------------------------------------------+
         |                                                                       |
         |                                                                       v
+----------------+                   +----------------+      Probe      +----------------+
|                |                   |                |<----------------|                |
|                |   Probe succeed   |                |                 |                |
|     Closed     |<------------------|    HalfOpen    |                 |      Open      |
|                |                   |                |   Probe failed  |                |
|                |                   |                +---------------->|                |
+----------------+                   +----------------+                 +----------------+
*/
type State int32

const (
	Closed State = iota
	HalfOpen
	Open
)

func (s *State) String() string {
	switch s.get() {
	case Closed:
		return "Closed"
	case HalfOpen:
		return "HalfOpen"
	case Open:
		return "Open"
	default:
		return "Undefined"
	}
}

func (s *State) get() State {
	statePtr := (*int32)(unsafe.Pointer(s))
	return State(atomic.LoadInt32(statePtr))
}

func (s *State) set(update State) {
	statePtr := (*int32)(unsafe.Pointer(s))
	newState := int32(update)
	atomic.StoreInt32(statePtr, newState)
}

func (s *State) casState(expect State, update State) bool {
	statePtr := (*int32)(unsafe.Pointer(s))
	oldState := int32(expect)
	newState := int32(update)
	return atomic.CompareAndSwapInt32(statePtr, oldState, newState)
}

// StateChangeListener listens on the circuit breaker state change event
type StateChangeListener interface {
	// OnTransformToClosed is triggered when circuit breaker state transformed to Closed.
	OnTransformToClosed(prev State, rule Rule)

	// OnTransformToOpen is triggered when circuit breaker state transformed to Open.
	// The "snapshot" indicates the triggered value when the transformation occurs.
	OnTransformToOpen(prev State, rule Rule, snapshot interface{})

	// OnTransformToHalfOpen is triggered when circuit breaker state transformed to HalfOpen.
	OnTransformToHalfOpen(prev State, rule Rule)
}

// CircuitBreaker is the basic interface of circuit breaker
type CircuitBreaker interface {
	// BoundRule returns the associated circuit breaking rule.
	BoundRule() Rule
	// BoundStat returns the associated statistic data structure.
	BoundStat() interface{}
	// TryPass acquires permission of an invocation only if it is available at the time of invocation.
	TryPass(ctx *base.EntryContext) bool
	// CurrentState returns current state of the circuit breaker.
	CurrentState() State
	// OnRequestComplete record a completed request with the given response time as well as error (if present),
	// and handle state transformation of the circuit breaker.
	OnRequestComplete(rtt uint64, err error)
}

//================================= circuitBreakerBase ====================================
// circuitBreakerBase encompasses the common fields of circuit breaker.
type circuitBreakerBase struct {
	rule Rule
	// retryTimeoutMs represents recovery timeout (in milliseconds) before the circuit breaker opens.
	// During the open period, no requests are permitted until the timeout has elapsed.
	// After that, the circuit breaker will transform to half-open state for trying a few "trial" requests.
	retryTimeoutMs uint32
	// nextRetryTimestampMs is the time circuit breaker could probe
	nextRetryTimestampMs uint64
	// state is the state machine of circuit breaker
	state *State
}

func (b *circuitBreakerBase) BoundRule() Rule {
	return b.rule
}

func (b *circuitBreakerBase) CurrentState() State {
	return b.state.get()
}

func (b *circuitBreakerBase) retryTimeoutArrived() bool {
	return util.CurrentTimeMillis() >= atomic.LoadUint64(&b.nextRetryTimestampMs)
}

func (b *circuitBreakerBase) updateNextRetryTimestamp() {
	atomic.StoreUint64(&b.nextRetryTimestampMs, util.CurrentTimeMillis()+uint64(b.retryTimeoutMs))
}

// fromClosedToOpen updates circuit breaker state machine from closed to open.
// Return true only if current goroutine successfully accomplished the transformation.
func (b *circuitBreakerBase) fromClosedToOpen(snapshot interface{}) bool {
	if b.state.casState(Closed, Open) {
		b.updateNextRetryTimestamp()
		for _, listener := range stateChangeListeners {
			listener.OnTransformToOpen(Closed, b.rule, snapshot)
		}
		return true
	}
	return false
}

// fromOpenToHalfOpen updates circuit breaker state machine from open to half-open.
// Return true only if current goroutine successfully accomplished the transformation.
func (b *circuitBreakerBase) fromOpenToHalfOpen() bool {
	if b.state.casState(Open, HalfOpen) {
		for _, listener := range stateChangeListeners {
			listener.OnTransformToHalfOpen(Open, b.rule)
		}
		return true
	}
	return false
}

// fromHalfOpenToOpen updates circuit breaker state machine from half-open to open.
// Return true only if current goroutine successfully accomplished the transformation.
func (b *circuitBreakerBase) fromHalfOpenToOpen(snapshot interface{}) bool {
	if b.state.casState(HalfOpen, Open) {
		b.updateNextRetryTimestamp()
		for _, listener := range stateChangeListeners {
			listener.OnTransformToOpen(HalfOpen, b.rule, snapshot)
		}
		return true
	}
	return false
}

// fromHalfOpenToOpen updates circuit breaker state machine from half-open to closed
// Return true only if current goroutine successfully accomplished the transformation.
func (b *circuitBreakerBase) fromHalfOpenToClosed() bool {
	if b.state.casState(HalfOpen, Closed) {
		for _, listener := range stateChangeListeners {
			listener.OnTransformToClosed(HalfOpen, b.rule)
		}
		return true
	}
	return false
}

//================================= slowRtCircuitBreaker ====================================
type slowRtCircuitBreaker struct {
	circuitBreakerBase
	stat                *slowRequestLeapArray
	maxAllowedRt        uint64
	maxSlowRequestRatio float64
	minRequestAmount    uint64
}

func newSlowRtCircuitBreakerWithStat(r *slowRtRule, stat *slowRequestLeapArray) *slowRtCircuitBreaker {
	status := new(State)
	status.set(Closed)
	return &slowRtCircuitBreaker{
		circuitBreakerBase: circuitBreakerBase{
			rule:                 r,
			retryTimeoutMs:       r.RetryTimeoutMs,
			nextRetryTimestampMs: 0,
			state:                status,
		},
		stat:                stat,
		maxAllowedRt:        r.MaxAllowedRtMs,
		maxSlowRequestRatio: r.MaxSlowRequestRatio,
		minRequestAmount:    r.MinRequestAmount,
	}
}

func newSlowRtCircuitBreaker(r *slowRtRule) *slowRtCircuitBreaker {
	interval := r.StatIntervalMs
	stat := &slowRequestLeapArray{}
	stat.data = sbase.NewLeapArray(1, interval, stat)

	return newSlowRtCircuitBreakerWithStat(r, stat)
}

func (b *slowRtCircuitBreaker) BoundStat() interface{} {
	return b.stat
}

// TryPass checks circuit breaker based on state machine of circuit breaker.
func (b *slowRtCircuitBreaker) TryPass(_ *base.EntryContext) bool {
	curStatus := b.CurrentState()
	if curStatus == Closed {
		return true
	} else if curStatus == Open {
		// switch state to half-open to probe if retry timeout
		if b.retryTimeoutArrived() && b.fromOpenToHalfOpen() {
			return true
		}
	}
	return false
}

func (b *slowRtCircuitBreaker) OnRequestComplete(rt uint64, err error) {
	// add slow and add total
	metricStat := b.stat
	counter := metricStat.currentCounter()
	if rt > b.maxAllowedRt {
		atomic.AddUint64(&counter.slowCount, 1)
	}
	atomic.AddUint64(&counter.totalCount, 1)

	slowCount := uint64(0)
	totalCount := uint64(0)
	counters := metricStat.allCounter()
	for _, c := range counters {
		slowCount += atomic.LoadUint64(&c.slowCount)
		totalCount += atomic.LoadUint64(&c.totalCount)
	}
	slowRatio := float64(slowCount) / float64(totalCount)

	// handleStateChange
	curStatus := b.CurrentState()
	if curStatus == Open {
		return
	} else if curStatus == HalfOpen {
		if rt > b.maxAllowedRt {
			// fail to probe
			b.fromHalfOpenToOpen(1.0)
		} else {
			// succeed to probe
			b.fromHalfOpenToClosed()
			b.resetMetric()
		}
		return
	}

	// current state is CLOSED
	if totalCount < b.minRequestAmount {
		return
	}

	if slowRatio > b.maxSlowRequestRatio {
		curStatus = b.CurrentState()
		switch curStatus {
		case Closed:
			b.fromClosedToOpen(slowRatio)
		case HalfOpen:
			b.fromHalfOpenToOpen(slowRatio)
		default:
		}
	}
	return
}

func (b *slowRtCircuitBreaker) resetMetric() {
	for _, c := range b.stat.allCounter() {
		c.reset()
	}
}

type slowRequestCounter struct {
	slowCount  uint64
	totalCount uint64
}

func (c *slowRequestCounter) reset() {
	atomic.StoreUint64(&c.slowCount, 0)
	atomic.StoreUint64(&c.totalCount, 0)
}

type slowRequestLeapArray struct {
	data *sbase.LeapArray
}

func (s *slowRequestLeapArray) NewEmptyBucket() interface{} {
	return &slowRequestCounter{
		slowCount:  0,
		totalCount: 0,
	}
}

func (s *slowRequestLeapArray) ResetBucketTo(bw *sbase.BucketWrap, startTime uint64) *sbase.BucketWrap {
	atomic.StoreUint64(&bw.BucketStart, startTime)
	bw.Value.Store(&slowRequestCounter{
		slowCount:  0,
		totalCount: 0,
	})
	return bw
}

func (s *slowRequestLeapArray) currentCounter() *slowRequestCounter {
	curBucket, err := s.data.CurrentBucket(s)
	if err != nil {
		logger.Errorf("Failed to get current bucket, current ts=%d, err: %+v.", util.CurrentTimeMillis(), err)
		return nil
	}
	if curBucket == nil {
		logger.Error("Current bucket is nil")
		return nil
	}
	mb := curBucket.Value.Load()
	if mb == nil {
		logger.Error("Current bucket atomic Value is nil")
		return nil
	}
	counter, ok := mb.(*slowRequestCounter)
	if !ok {
		logger.Error("Bucket data type error")
		return nil
	}
	return counter
}

func (s *slowRequestLeapArray) allCounter() []*slowRequestCounter {
	buckets := s.data.Values()
	ret := make([]*slowRequestCounter, 0)
	for _, b := range buckets {
		mb := b.Value.Load()
		if mb == nil {
			logger.Error("Current bucket atomic Value is nil")
			continue
		}
		counter, ok := mb.(*slowRequestCounter)
		if !ok {
			logger.Error("Bucket data type error")
			continue
		}
		ret = append(ret, counter)
	}
	return ret
}

//================================= errorRatioCircuitBreaker ====================================
type errorRatioCircuitBreaker struct {
	circuitBreakerBase
	minRequestAmount    uint64
	errorRatioThreshold float64

	stat *errorCounterLeapArray
}

func newErrorRatioCircuitBreakerWithStat(r *errorRatioRule, stat *errorCounterLeapArray) *errorRatioCircuitBreaker {
	status := new(State)
	status.set(Closed)

	return &errorRatioCircuitBreaker{
		circuitBreakerBase: circuitBreakerBase{
			rule:                 r,
			retryTimeoutMs:       r.RetryTimeoutMs,
			nextRetryTimestampMs: 0,
			state:                status,
		},
		minRequestAmount:    r.MinRequestAmount,
		errorRatioThreshold: r.Threshold,
		stat:                stat,
	}
}

func newErrorRatioCircuitBreaker(r *errorRatioRule) *errorRatioCircuitBreaker {
	interval := r.StatIntervalMs
	stat := &errorCounterLeapArray{}
	stat.data = sbase.NewLeapArray(1, interval, stat)

	return newErrorRatioCircuitBreakerWithStat(r, stat)
}

func (b *errorRatioCircuitBreaker) BoundStat() interface{} {
	return b.stat
}

func (b *errorRatioCircuitBreaker) TryPass(_ *base.EntryContext) bool {
	curStatus := b.CurrentState()
	if curStatus == Closed {
		return true
	} else if curStatus == Open {
		// switch state to half-open to probe if retry timeout
		if b.retryTimeoutArrived() && b.fromOpenToHalfOpen() {
			return true
		}
	}
	return false
}

func (b *errorRatioCircuitBreaker) OnRequestComplete(rt uint64, err error) {
	metricStat := b.stat
	counter := metricStat.currentCounter()
	if err != nil {
		atomic.AddUint64(&counter.errorCount, 1)
	}
	atomic.AddUint64(&counter.totalCount, 1)

	errorCount := uint64(0)
	totalCount := uint64(0)
	counters := metricStat.allCounter()
	for _, c := range counters {
		errorCount += atomic.LoadUint64(&c.errorCount)
		totalCount += atomic.LoadUint64(&c.totalCount)
	}
	errorRatio := float64(errorCount) / float64(totalCount)

	// handleStateChangeWhenThresholdExceeded
	curStatus := b.CurrentState()
	if curStatus == Open {
		return
	}
	if curStatus == HalfOpen {
		if err == nil {
			b.fromHalfOpenToClosed()
			b.resetMetric()
		} else {
			b.fromHalfOpenToOpen(1.0)
		}
		return
	}

	// current state is CLOSED
	if totalCount < b.minRequestAmount {
		return
	}
	if errorRatio > b.errorRatioThreshold {
		curStatus = b.CurrentState()
		switch curStatus {
		case Closed:
			b.fromClosedToOpen(errorRatio)
		case HalfOpen:
			b.fromHalfOpenToOpen(errorRatio)
		default:
		}
	}
}

func (b *errorRatioCircuitBreaker) resetMetric() {
	for _, c := range b.stat.allCounter() {
		c.reset()
	}
}

type errorCounter struct {
	errorCount uint64
	totalCount uint64
}

func (c *errorCounter) reset() {
	atomic.StoreUint64(&c.errorCount, 0)
	atomic.StoreUint64(&c.totalCount, 0)
}

type errorCounterLeapArray struct {
	data *sbase.LeapArray
}

func (s *errorCounterLeapArray) NewEmptyBucket() interface{} {
	return &errorCounter{
		errorCount: 0,
		totalCount: 0,
	}
}

func (s *errorCounterLeapArray) ResetBucketTo(bw *sbase.BucketWrap, startTime uint64) *sbase.BucketWrap {
	atomic.StoreUint64(&bw.BucketStart, startTime)
	bw.Value.Store(&errorCounter{
		errorCount: 0,
		totalCount: 0,
	})
	return bw
}

func (s *errorCounterLeapArray) currentCounter() *errorCounter {
	curBucket, err := s.data.CurrentBucket(s)
	if err != nil {
		logger.Errorf("Failed to get current bucket, current ts=%d, err: %+v.", util.CurrentTimeMillis(), err)
		return nil
	}
	if curBucket == nil {
		logger.Error("Current bucket is nil")
		return nil
	}
	mb := curBucket.Value.Load()
	if mb == nil {
		logger.Error("Current bucket atomic Value is nil")
		return nil
	}
	counter, ok := mb.(*errorCounter)
	if !ok {
		logger.Error("Bucket data type error")
		return nil
	}
	return counter
}

func (s *errorCounterLeapArray) allCounter() []*errorCounter {
	buckets := s.data.Values()
	ret := make([]*errorCounter, 0)
	for _, b := range buckets {
		mb := b.Value.Load()
		if mb == nil {
			logger.Error("Current bucket atomic Value is nil")
			continue
		}
		counter, ok := mb.(*errorCounter)
		if !ok {
			logger.Error("Bucket data type error")
			continue
		}
		ret = append(ret, counter)
	}
	return ret
}

//================================= errorCountCircuitBreaker ====================================
type errorCountCircuitBreaker struct {
	circuitBreakerBase
	minRequestAmount    uint64
	errorCountThreshold uint64

	stat *errorCounterLeapArray
}

func newErrorCountCircuitBreakerWithStat(r *errorCountRule, stat *errorCounterLeapArray) *errorCountCircuitBreaker {
	status := new(State)
	status.set(Closed)

	return &errorCountCircuitBreaker{
		circuitBreakerBase: circuitBreakerBase{
			rule:                 r,
			retryTimeoutMs:       r.RetryTimeoutMs,
			nextRetryTimestampMs: 0,
			state:                status,
		},
		minRequestAmount:    r.MinRequestAmount,
		errorCountThreshold: r.Threshold,
		stat:                stat,
	}
}

func newErrorCountCircuitBreaker(r *errorCountRule) *errorCountCircuitBreaker {
	interval := r.StatIntervalMs
	stat := &errorCounterLeapArray{}
	stat.data = sbase.NewLeapArray(1, interval, stat)

	return newErrorCountCircuitBreakerWithStat(r, stat)
}

func (b *errorCountCircuitBreaker) BoundStat() interface{} {
	return b.stat
}

func (b *errorCountCircuitBreaker) TryPass(_ *base.EntryContext) bool {
	curStatus := b.CurrentState()
	if curStatus == Closed {
		return true
	} else if curStatus == Open {
		// switch state to half-open to probe if retry timeout
		if b.retryTimeoutArrived() && b.fromOpenToHalfOpen() {
			return true
		}
	}
	return false
}

func (b *errorCountCircuitBreaker) OnRequestComplete(rt uint64, err error) {
	metricStat := b.stat
	counter := metricStat.currentCounter()
	if err != nil {
		atomic.AddUint64(&counter.errorCount, 1)
	}
	atomic.AddUint64(&counter.totalCount, 1)

	errorCount := uint64(0)
	totalCount := uint64(0)
	counters := metricStat.allCounter()
	for _, c := range counters {
		errorCount += atomic.LoadUint64(&c.errorCount)
		totalCount += atomic.LoadUint64(&c.totalCount)
	}
	// handleStateChangeWhenThresholdExceeded
	curStatus := b.CurrentState()
	if curStatus == Open {
		return
	}
	if curStatus == HalfOpen {
		if err == nil {
			b.fromHalfOpenToClosed()
			b.resetMetric()
		} else {
			b.fromHalfOpenToOpen(1)
		}
		return
	}
	// current state is CLOSED
	if totalCount < b.minRequestAmount {
		return
	}
	if errorCount > b.errorCountThreshold {
		curStatus = b.CurrentState()
		switch curStatus {
		case Closed:
			b.fromClosedToOpen(errorCount)
		case HalfOpen:
			b.fromHalfOpenToOpen(errorCount)
		default:
		}
	}
}

func (b *errorCountCircuitBreaker) resetMetric() {
	for _, c := range b.stat.allCounter() {
		c.reset()
	}
}
