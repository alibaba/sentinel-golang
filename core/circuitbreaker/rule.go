package circuitbreaker

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// Strategy represents the strategy of circuit breaker.
// Each strategy is associated with one rule type.
type Strategy int8

const (
	// SlowRequestRatio strategy changes the circuit breaker state based on slow request ratio
	SlowRequestRatio Strategy = iota
	// ErrorRatio strategy changes the circuit breaker state based on error request ratio
	ErrorRatio
	// ErrorCount strategy changes the circuit breaker state based on error amount
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

// Rule is the base interface of the circuit breaker rule.
type Rule interface {
	base.SentinelRule
	// BreakerStrategy returns the circuit breaker strategy.
	BreakerStrategy() Strategy
	// IsApplicable checks whether the rule is valid and could be converted to a corresponding circuit breaker.
	IsApplicable() error
	// BreakerStatIntervalMs returns the statistic interval of circuit breaker (in milliseconds).
	BreakerStatIntervalMs() uint32
	// IsEqualsTo checks whether current rule is equal to the given rule.
	IsEqualsTo(r Rule) bool
	// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
	IsStatReusable(r Rule) bool
}

// RuleBase encompasses the common fields of circuit breaking rule.
type RuleBase struct {
	// unique id
	Id string
	// resource name
	Resource string
	Strategy Strategy
	// RetryTimeoutMs represents recovery timeout (in milliseconds) before the circuit breaker opens.
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
	if len(b.Resource) == 0 {
		return errors.New("empty resource name")
	}
	if b.RetryTimeoutMs <= 0 {
		return errors.New("invalid RetryTimeoutMs")
	}
	if b.MinRequestAmount <= 0 {
		return errors.New("invalid MinRequestAmount")
	}
	if b.StatIntervalMs <= 0 {
		return errors.New("invalid StatIntervalMs")
	}
	return nil
}

func (b *RuleBase) IsStatReusable(r Rule) bool {
	return b.Resource == r.ResourceName() && b.Strategy == r.BreakerStrategy() && b.StatIntervalMs == r.BreakerStatIntervalMs()
}

func (b *RuleBase) String() string {
	// fallback string
	return fmt.Sprintf("{id=%s,resource=%s, strategy=%+v, RetryTimeoutMs=%d, MinRequestAmount=%d, StatIntervalMs=%d}",
		b.Id, b.Resource, b.Strategy, b.RetryTimeoutMs, b.MinRequestAmount, b.StatIntervalMs)
}

func (b *RuleBase) BreakerStrategy() Strategy {
	return b.Strategy
}

func (b *RuleBase) ResourceName() string {
	return b.Resource
}

type RuleOptions struct {
	retryTimeoutMs   uint32
	minRequestAmount uint64
	statIntervalMs   uint32

	//The following two fields apply only to slowRtRule
	maxAllowedRtMs      uint64
	maxSlowRequestRatio float64

	//The following one field apply only to errorRatioRule
	errorRatioThreshold float64

	//The following one field apply only to errorCountRule
	errorCountThreshold uint64
}

// TODO: make default option configurable?
func newDefaultRuleOptions() *RuleOptions {
	return &RuleOptions{
		retryTimeoutMs:      0,
		minRequestAmount:    0,
		statIntervalMs:      0,
		maxAllowedRtMs:      0,
		maxSlowRequestRatio: 0,
		errorRatioThreshold: 0,
		errorCountThreshold: 0,
	}
}

type RuleOption func(opts *RuleOptions)

// WithRetryTimeoutMs sets the retryTimeoutMs
// This function takes effect for all circuit breaker rule
func WithRetryTimeoutMs(retryTimeoutMs uint32) RuleOption {
	return func(opts *RuleOptions) {
		opts.retryTimeoutMs = retryTimeoutMs
	}
}

// WithMinRequestAmount sets the minRequestAmount
// This function takes effect for all circuit breaker rule
func WithMinRequestAmount(minRequestAmount uint64) RuleOption {
	return func(opts *RuleOptions) {
		opts.minRequestAmount = minRequestAmount
	}
}

// WithStatIntervalMs sets the statIntervalMs
// This function takes effect for all circuit breaker rule
func WithStatIntervalMs(statIntervalMs uint32) RuleOption {
	return func(opts *RuleOptions) {
		opts.statIntervalMs = statIntervalMs
	}
}

// WithMaxAllowedRtMs sets the maxAllowedRtMs
// This function only takes effect for slowRtRule
func WithMaxAllowedRtMs(maxAllowedRtMs uint64) RuleOption {
	return func(opts *RuleOptions) {
		opts.maxAllowedRtMs = maxAllowedRtMs
	}
}

// WithMaxSlowRequestRatio sets the maxSlowRequestRatio
// This function only takes effect for slowRtRule
func WithMaxSlowRequestRatio(maxSlowRequestRatio float64) RuleOption {
	return func(opts *RuleOptions) {
		opts.maxSlowRequestRatio = maxSlowRequestRatio
	}
}

// WithErrorRatioThreshold sets the errorRatioThreshold
// This function only takes effect for errorRatioRule
func WithErrorRatioThreshold(errorRatioThreshold float64) RuleOption {
	return func(opts *RuleOptions) {
		opts.errorRatioThreshold = errorRatioThreshold
	}
}

// WithErrorCountThreshold sets the errorCountThreshold
// This function only takes effect for errorCountRule
func WithErrorCountThreshold(errorCountThreshold uint64) RuleOption {
	return func(opts *RuleOptions) {
		opts.errorCountThreshold = errorCountThreshold
	}
}

// SlowRequestRatio circuit breaker rule
type slowRtRule struct {
	RuleBase
	// MaxAllowedRtMs indicates that any invocation whose response time exceeds this value (in ms)
	// will be recorded as a slow request.
	MaxAllowedRtMs uint64
	// MaxSlowRequestRatio represents the threshold of slow rt ratio circuit breaker.
	MaxSlowRequestRatio float64
}

func NewRule(resource string, strategy Strategy, opts ...RuleOption) Rule {
	ruleOpts := newDefaultRuleOptions()
	for _, opt := range opts {
		opt(ruleOpts)
	}

	switch strategy {
	case SlowRequestRatio:
		return &slowRtRule{
			RuleBase: RuleBase{
				Id:               util.NewUuid(),
				Resource:         resource,
				Strategy:         SlowRequestRatio,
				RetryTimeoutMs:   ruleOpts.retryTimeoutMs,
				MinRequestAmount: ruleOpts.minRequestAmount,
				StatIntervalMs:   ruleOpts.statIntervalMs,
			},
			MaxAllowedRtMs:      ruleOpts.maxAllowedRtMs,
			MaxSlowRequestRatio: ruleOpts.maxSlowRequestRatio,
		}
	case ErrorRatio:
		return &errorRatioRule{
			RuleBase: RuleBase{
				Id:               util.NewUuid(),
				Resource:         resource,
				Strategy:         ErrorRatio,
				RetryTimeoutMs:   ruleOpts.retryTimeoutMs,
				MinRequestAmount: ruleOpts.minRequestAmount,
				StatIntervalMs:   ruleOpts.statIntervalMs,
			},
			Threshold: ruleOpts.errorRatioThreshold,
		}
	case ErrorCount:
		return &errorCountRule{
			RuleBase: RuleBase{
				Id:               util.NewUuid(),
				Resource:         resource,
				Strategy:         ErrorCount,
				RetryTimeoutMs:   ruleOpts.retryTimeoutMs,
				MinRequestAmount: ruleOpts.minRequestAmount,
				StatIntervalMs:   ruleOpts.statIntervalMs,
			},
			Threshold: ruleOpts.errorCountThreshold,
		}
	default:
		logger.Errorf("unsupported circuit breaker rule, strategy: %d", strategy)
		return nil
	}
}

func (r *slowRtRule) IsEqualsTo(newRule Rule) bool {
	newSlowRtRule, ok := newRule.(*slowRtRule)
	if !ok {
		return false
	}
	return r.Resource == newSlowRtRule.Resource && r.Strategy == newSlowRtRule.Strategy && r.RetryTimeoutMs == newSlowRtRule.RetryTimeoutMs &&
		r.MinRequestAmount == newSlowRtRule.MinRequestAmount && r.StatIntervalMs == newSlowRtRule.StatIntervalMs &&
		r.MaxAllowedRtMs == newSlowRtRule.MaxAllowedRtMs && r.MaxSlowRequestRatio == newSlowRtRule.MaxSlowRequestRatio
}

func (r *slowRtRule) IsApplicable() error {
	baseCheckErr := r.RuleBase.IsApplicable()
	if baseCheckErr != nil {
		return baseCheckErr
	}
	if r.MaxSlowRequestRatio < 0 || r.MaxSlowRequestRatio > 1 {
		return errors.New("invalid slow request ratio threshold (valid range: [0.0, 1.0])")
	}
	return nil
}

func (r *slowRtRule) String() string {
	return fmt.Sprintf("{slowRtRule{RuleBase:%s, MaxAllowedRtMs=%d, MaxSlowRequestRatio=%f}", r.RuleBase.String(), r.MaxAllowedRtMs, r.MaxSlowRequestRatio)
}

// Error ratio circuit breaker rule
type errorRatioRule struct {
	RuleBase
	Threshold float64
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
	baseCheckErr := r.RuleBase.IsApplicable()
	if baseCheckErr != nil {
		return baseCheckErr
	}
	if r.Threshold < 0 || r.Threshold > 1 {
		return errors.New("invalid error ratio threshold (valid range: [0.0, 1.0])")
	}
	return nil
}

// Error count circuit breaker rule
type errorCountRule struct {
	RuleBase
	Threshold uint64
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
	baseCheckErr := r.RuleBase.IsApplicable()
	if baseCheckErr != nil {
		return baseCheckErr
	}
	if r.Threshold < 0 {
		return errors.New("negative error count threshold")
	}
	return nil
}
