package circuit_breaker

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// The strategy of circuit breaker
// Each strategy represent one rule type
type BreakerStrategy int8

const (
	AverageRt BreakerStrategy = iota
	ErrorRatio
	ErrorCount
)

// The base interface of circuit breaker rule
type Rule interface {
	base.SentinelRule
	// return the strategy type
	BreakerStrategy() BreakerStrategy
	// check whether the rule is valid and could be converted to corresponding circuit breaker
	isApplicable() bool
	// convert circuit breaker rule to circuit breaker
	convert2CircuitBreaker() CircuitBreaker
}

// The common fields of circuit breaker rule
type ruleBase struct {
	// unique id
	Id string `json:"id,omitempty"`
	// resource name
	Resource string          `json:"resource"`
	Strategy BreakerStrategy `json:"strategy"`
	// auto recover timeout in second, all requests would be broken before auto recover
	RecoverTimeout int64 `json:"recoverTimeout"`
	// the base data to describe the statistic metric
	SampleCount  uint32 `json:"sampleCount"`
	IntervalInMs uint32 `json:"intervalInMs"`
}

func (b *ruleBase) isApplicable() bool {
	if !(len(b.Resource) > 0 && b.RecoverTimeout >= 0) {
		logger.Warnf("Illegal parameters,Resource=%s,RecoverTimeout=%d.", b.Resource, b.RecoverTimeout)
		return false
	}
	if b.IntervalInMs <= 0 || b.SampleCount <= 0 {
		logger.Warnf("Illegal parameters,SampleCount=%d,IntervalInMs=%d.", b.SampleCount, b.IntervalInMs)
		return false
	}

	if b.IntervalInMs%b.SampleCount != 0 {
		logger.Warnf("Invalid parameters, SampleCount=%d,IntervalInMs=%d.", b.SampleCount, b.IntervalInMs)
		return false
	}
	return true
}

func (b *ruleBase) String() string {
	r, err := json.Marshal(b)
	if err != nil {
		// fallback string
		return fmt.Sprintf("ruleBase{id=%s,resource=%s, strategy=%d, RecoverTimeout=%d, SampleCount=%d, IntervalInMs=%d}, err:%+v.",
			b.Id, b.Resource, b.Strategy, b.RecoverTimeout, b.SampleCount, b.IntervalInMs, errors.WithStack(err))
	}
	return string(r)
}

func (b *ruleBase) BreakerStrategy() BreakerStrategy {
	return b.Strategy
}

func (b *ruleBase) ResourceName() string {
	return b.Resource
}

// Average Rt circuit breaker rule
type averageRtRule struct {
	ruleBase
	// the threshold of rt(ms)
	Threshold float64 `json:"threshold"`
	// if average rt > threshold && the count of request exceed RtSlowRequestAmount, then trigger circuit breaker
	RtSlowRequestAmount int64 `json:"rtSlowRequestAmount"`
}

func NewAverageRtRule(resource string, recoverTimeout int64, sampleCount, intervalInMs uint32, threshold float64, rtSlowRequestAmount int64) *averageRtRule {
	return &averageRtRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       resource,
			Strategy:       AverageRt,
			RecoverTimeout: recoverTimeout,
			SampleCount:    sampleCount,
			IntervalInMs:   intervalInMs,
		},
		Threshold:           threshold,
		RtSlowRequestAmount: rtSlowRequestAmount,
	}
}

func (r *averageRtRule) isApplicable() bool {
	if !r.ruleBase.isApplicable() {
		return false
	}
	if !(r.BreakerStrategy() == AverageRt && r.Threshold >= 0.0 && r.RtSlowRequestAmount >= 0) {
		return false
	}
	return true
}

func (r *averageRtRule) String() string {
	ret, err := json.Marshal(r)
	if err != nil {
		// feedback string
		return fmt.Sprintf("averageRtRule{ruleBase:%s, threshold=%f,rRtSlowRequestAmount=%d}, err:%+v.",
			r.ruleBase.String(), r.Threshold, r.RtSlowRequestAmount, errors.WithStack(err))
	}
	return string(ret)
}

func (r *averageRtRule) convert2CircuitBreaker() CircuitBreaker {
	return newAverageRtCircuitBreaker(r)
}

// Error ratio circuit breaker rule
type errorRatioRule struct {
	ruleBase
	Threshold float64 `json:"threshold"`
	// if request count < MinRequestAmount, pass the rule checker directly.
	MinRequestAmount int64 `json:"minRequestAmount"`
}

func NewErrorRatioRule(resource string, recoverTimeout int64, sampleCount, intervalInMs uint32, threshold float64, rtSlowRequestAmount int64) *errorRatioRule {
	return &errorRatioRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       resource,
			Strategy:       ErrorRatio,
			RecoverTimeout: recoverTimeout,
			SampleCount:    sampleCount,
			IntervalInMs:   intervalInMs,
		},
		Threshold:        threshold,
		MinRequestAmount: rtSlowRequestAmount,
	}
}

func (r *errorRatioRule) String() string {
	ret, err := json.Marshal(r)
	if err != nil {
		// feedback string
		return fmt.Sprintf("errorRatioRule{ruleBase:%s, threshold=%f, minRequestAmount=%d}, err:%+v.",
			r.ruleBase.String(), r.Threshold, r.MinRequestAmount, errors.WithStack(err))
	}
	return string(ret)
}

func (r *errorRatioRule) isApplicable() bool {
	if !r.ruleBase.isApplicable() {
		return false
	}
	if !(r.BreakerStrategy() == ErrorRatio && r.Threshold >= float64(0.0) && r.MinRequestAmount >= 0) {
		return false
	}
	return true
}

func (r *errorRatioRule) convert2CircuitBreaker() CircuitBreaker {
	return newErrorRatioCircuitBreaker(r)
}

// Error count circuit breaker rule
type errorCountRule struct {
	ruleBase
	Threshold int64 `json:"threshold"`
}

func NewErrorCountRule(resource string, recoverTimeout int64, sampleCount, intervalInMs uint32, threshold int64) *errorCountRule {
	return &errorCountRule{
		ruleBase: ruleBase{
			Id:             util.NewUuid(),
			Resource:       resource,
			Strategy:       ErrorCount,
			RecoverTimeout: recoverTimeout,
			SampleCount:    sampleCount,
			IntervalInMs:   intervalInMs,
		},
		Threshold: threshold,
	}
}

func (r *errorCountRule) String() string {
	ret, err := json.Marshal(r)
	if err != nil {
		// feedback string
		return fmt.Sprintf("errorCountRule{ruleBase:%s, threshold=%d} err:%+v.",
			r.ruleBase.String(), r.Threshold, errors.WithStack(err))
	}
	return string(ret)
}

func (r *errorCountRule) isApplicable() bool {
	if !r.ruleBase.isApplicable() {
		return false
	}
	if !(r.BreakerStrategy() == ErrorCount && r.Threshold >= 0) {
		return false
	}
	return true
}

func (r *errorCountRule) convert2CircuitBreaker() CircuitBreaker {
	return newErrorCountCircuitBreaker(r)
}
