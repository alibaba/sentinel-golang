package circuitbreaker

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// The strategy of circuit breaker.
// Each strategy represents one rule type.
type Strategy int8

const (
	SlowRequestRatio Strategy = iota
	ErrorRatio
	ErrorCount
)

func (s Strategy) String() string {
	switch s {
	case SlowRequestRatio:
		return "SlowRequestRatio"
	case ErrorRatio:
		return "ErrorRatio"
	case ErrorCount:
		return "ErrorCount"
	default:
		return "Undefined"
	}
}

// Rule represents the base interface of the circuit breaker rule.
type Rule interface {
	base.SentinelRule
	// BreakerStrategy returns the strategy.
	BreakerStrategy() Strategy
	// IsApplicable checks whether the rule is valid and could be converted to a corresponding circuit breaker.
	IsApplicable() error
	// BreakerStatIntervalMs returns the statistic interval of circuit breaker (in milliseconds).
	BreakerStatIntervalMs() uint32
	// IsEqualsTo checks whether current rule is consistent with the given rule.
	IsEqualsTo(r Rule) bool
	// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
	IsStatReusable(r Rule) bool
}

// RuleBase encompasses common fields of circuit breaking rule.
type RuleBase struct {
	// unique id
	Id string
	// resource name
	Resource string
	Strategy Strategy
	// RetryTimeoutMs represents recovery timeout (in seconds) before the circuit breaker opens.
	// During the open period, no requests are permitted until the timeout has elapsed.
	// After that, the circuit breaker will transform to half-open state for trying a few "trial" requests.
	RetryTimeoutMs uint32
	// MinRequestAmount represents the minimum number of requests (in an active statistic time span)
	// that can trigger circuit breaking.
	MinRequestAmount uint64
	// StatIntervalMs represents statistic time interval of the internal circuit breaker (in ms).
	StatIntervalMs uint32
}

func (b *RuleBase) BreakerStatIntervalMs() uint32 {
	return b.StatIntervalMs
}

func (b *RuleBase) IsApplicable() error {
	if !(len(b.Resource) > 0 && b.RetryTimeoutMs >= 0 && b.MinRequestAmount >= 0 && b.StatIntervalMs >= 0) {
		return errors.Errorf("Illegal parameters, Id=%s, Resource=%s, Strategy=%d, RetryTimeoutMs=%d, MinRequestAmount=%d, StatIntervalMs=%d.",
			b.Id, b.Resource, b.Strategy, b.RetryTimeoutMs, b.MinRequestAmount, b.StatIntervalMs)
	}
	return nil
}

func (b *RuleBase) IsStatReusable(r Rule) bool {
	return b.Resource == r.ResourceName() && b.Strategy == r.BreakerStrategy() && b.StatIntervalMs == r.BreakerStatIntervalMs()
}

func (b *RuleBase) String() string {
	// fallback string
	return fmt.Sprintf("{id=%s,resource=%s, strategy=%+v, RetryTimeoutMs=%d, MinRequestAmount=%d}",
		b.Id, b.Resource, b.Strategy, b.RetryTimeoutMs, b.MinRequestAmount)
}

func (b *RuleBase) BreakerStrategy() Strategy {
	return b.Strategy
}

func (b *RuleBase) ResourceName() string {
	return b.Resource
}

// SlowRequestRatio circuit breaker rule
type slowRtRule struct {
	RuleBase
	// MaxAllowedRt indicates that any invocation whose response time exceeds this value
	// will be recorded as a slow request.
	MaxAllowedRt uint64
	// MaxSlowRequestRatio represents the threshold of slow request ratio.
	MaxSlowRequestRatio float64
}

func NewSlowRtRule(resource string, intervalMs uint32, retryTimeoutMs uint32, maxAllowedRt, minRequestAmount uint64, maxSlowRequestRatio float64) *slowRtRule {
	return &slowRtRule{
		RuleBase: RuleBase{
			Id:               util.NewUuid(),
			Resource:         resource,
			Strategy:         SlowRequestRatio,
			RetryTimeoutMs:   retryTimeoutMs,
			MinRequestAmount: minRequestAmount,
			StatIntervalMs:   intervalMs,
		},
		MaxAllowedRt:        maxAllowedRt,
		MaxSlowRequestRatio: maxSlowRequestRatio,
	}
}

func (r *slowRtRule) IsEqualsTo(newRule Rule) bool {
	newSlowRtRule, ok := newRule.(*slowRtRule)
	if !ok {
		return false
	}
	return r.Resource == newSlowRtRule.Resource && r.Strategy == newSlowRtRule.Strategy && r.RetryTimeoutMs == newSlowRtRule.RetryTimeoutMs &&
		r.MinRequestAmount == newSlowRtRule.MinRequestAmount && r.StatIntervalMs == newSlowRtRule.StatIntervalMs &&
		r.MaxAllowedRt == newSlowRtRule.MaxAllowedRt && r.MaxSlowRequestRatio == newSlowRtRule.MaxSlowRequestRatio
}

func (r *slowRtRule) IsApplicable() error {
	baseApplicableError := r.RuleBase.IsApplicable()
	var slowRtError error
	if !(r.MaxSlowRequestRatio >= 0.0 && r.MaxAllowedRt >= 0) {
		slowRtError = errors.Errorf("Illegal parameters in slowRtRule, MaxSlowRequestRatio: %f, MaxAllowedRt: %d", r.MaxSlowRequestRatio, r.MaxAllowedRt)
	}
	return multierr.Append(baseApplicableError, slowRtError)
}

func (r *slowRtRule) String() string {
	return fmt.Sprintf("{slowRtRule{RuleBase:%s, MaxAllowedRt=%d, MaxSlowRequestRatio=%f}", r.RuleBase.String(), r.MaxAllowedRt, r.MaxSlowRequestRatio)
}

// Error ratio circuit breaker rule
type errorRatioRule struct {
	RuleBase
	Threshold float64
}

func NewErrorRatioRule(resource string, intervalMs uint32, retryTimeoutMs uint32, minRequestAmount uint64, maxErrorRatio float64) *errorRatioRule {
	return &errorRatioRule{
		RuleBase: RuleBase{
			Id:               util.NewUuid(),
			Resource:         resource,
			Strategy:         ErrorRatio,
			RetryTimeoutMs:   retryTimeoutMs,
			MinRequestAmount: minRequestAmount,
			StatIntervalMs:   intervalMs,
		},
		Threshold: maxErrorRatio,
	}
}

func (r *errorRatioRule) String() string {
	return fmt.Sprintf("{errorRatioRule{RuleBase:%s, Threshold=%f}", r.RuleBase.String(), r.Threshold)
}

func (r *errorRatioRule) IsEqualsTo(newRule Rule) bool {
	newErrorRatioRule, ok := newRule.(*errorRatioRule)
	if !ok {
		return false
	}
	return r.Resource == newErrorRatioRule.Resource && r.Strategy == newErrorRatioRule.Strategy && r.RetryTimeoutMs == newErrorRatioRule.RetryTimeoutMs &&
		r.MinRequestAmount == newErrorRatioRule.MinRequestAmount && r.StatIntervalMs == newErrorRatioRule.StatIntervalMs &&
		r.Threshold == newErrorRatioRule.Threshold
}

func (r *errorRatioRule) IsApplicable() error {
	baseApplicableError := r.RuleBase.IsApplicable()
	var errorRatioRuleError error
	if !(r.Threshold >= 0.0) {
		errorRatioRuleError = errors.Errorf("Illegal parameters in errorRatioRule, Threshold: %f.", r.Threshold)
	}
	return multierr.Append(baseApplicableError, errorRatioRuleError)
}

// Error count circuit breaker rule
type errorCountRule struct {
	RuleBase
	Threshold uint64
}

func NewErrorCountRule(resource string, intervalMs uint32, retryTimeoutMs uint32, minRequestAmount, maxErrorCount uint64) *errorCountRule {
	return &errorCountRule{
		RuleBase: RuleBase{
			Id:               util.NewUuid(),
			Resource:         resource,
			Strategy:         ErrorCount,
			RetryTimeoutMs:   retryTimeoutMs,
			MinRequestAmount: minRequestAmount,
			StatIntervalMs:   intervalMs,
		},
		Threshold: maxErrorCount,
	}
}

func (r *errorCountRule) String() string {
	return fmt.Sprintf("{errorCountRule{RuleBase:%s, Threshold=%d}", r.RuleBase.String(), r.Threshold)
}

func (r *errorCountRule) IsEqualsTo(newRule Rule) bool {
	newErrorCountRule, ok := newRule.(*errorCountRule)
	if !ok {
		return false
	}
	return r.Resource == newErrorCountRule.Resource && r.Strategy == newErrorCountRule.Strategy && r.RetryTimeoutMs == newErrorCountRule.RetryTimeoutMs &&
		r.MinRequestAmount == newErrorCountRule.MinRequestAmount && r.StatIntervalMs == newErrorCountRule.StatIntervalMs &&
		r.Threshold == newErrorCountRule.Threshold
}

func (r *errorCountRule) IsApplicable() error {
	baseApplicableError := r.RuleBase.IsApplicable()
	var errorCountRuleError error
	if !(r.Threshold >= 0) {
		errorCountRuleError = errors.Errorf("Illegal parameters in errorCountRule, Threshold: %d.", r.Threshold)
	}
	return multierr.Append(baseApplicableError, errorCountRuleError)
}
