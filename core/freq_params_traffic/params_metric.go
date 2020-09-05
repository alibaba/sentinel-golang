package freq_params_traffic

import "github.com/alibaba/sentinel-golang/core/freq_params_traffic/cache"

const (
	ConcurrencyMaxCount = 4000
	ParamsCapacityBase  = 4000
	ParamsMaxCapacity   = 20000
)

// ParamsMetric cache the frequent(hot spot) parameters for each value.
// ParamsMetric is used for pair <resource, TrafficShapingController>.
type ParamsMetric struct {
	// cache's key is the hot value
	// cache's value is the counter
	// RuleTimeCounter record the last add token time
	RuleTimeCounter cache.ConcurrentCounterCache
	// RuleTokenCounter record the number of token
	RuleTokenCounter cache.ConcurrentCounterCache
	// ConcurrencyCounter record the number of goroutine
	ConcurrencyCounter cache.ConcurrentCounterCache
}
