package base

type TimePredicate func(uint64) bool

type MetricEvent int8

// There are five events to record
// pass + block == Total
const (
	// sentinel rules check pass
	MetricEventPass MetricEvent = iota
	// sentinel rules check block
	MetricEventBlock

	MetricEventComplete
	// Biz error, used for circuit breaker
	MetricEventError
	// request execute rt, unit is millisecond
	MetricEventRt
	// hack for the number of event
	MetricEventTotal
)

type StatGetter interface {
	GetSum(e MetricEvent) uint64
	GetAvg(e MetricEvent) float64
}

type StatUpdater interface {
	AddMetric(e MetricEvent, count uint64)
}

type ReadStat interface {
	GetQPS(event MetricEvent) float64
	GetQPSWithTime(now uint64, event MetricEvent) float64

	GetSum(event MetricEvent) int64
	GetSumWithTime(now uint64, event MetricEvent) int64

	MinRT() int64
	AvgRT() float64
}

type WriteStat interface {
	AddRequest(event MetricEvent, count uint64)
	AddRtAndCompleteRequest(rt, count uint64)
}

// StatNode holds real-time statistics for resources.
type StatNode interface {
	MetricItemRetriever

	ReadStat
	WriteStat

	CurrentGoroutineNum() int32
	IncreaseGoroutineNum()
	DecreaseGoroutineNum()

	Reset()
}
