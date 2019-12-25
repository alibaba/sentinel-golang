package core

type TimePredicate func(uint64) bool
type MetricEvent int32

const (
	MetricEventPass MetricEvent = iota
	MetricEventBlock
	MetricEventComplete
	MetricEventError
	MetricEventRT
)

type StatGetter interface {
	GetSum(e MetricEvent) uint64
	GetAvg(e MetricEvent) float64
}

type StatUpdater interface {
	AddMetric(e MetricEvent, count uint64)
}

// StatNode holds real-time statistics for resources.
type StatNode interface {
	MetricItemRetriever

	// total  = pass + blocked
	TotalQPS() float64
	PassQPS() float64
	BlockQPS() float64
	CompleteQPS() float64
	ErrorQPS() float64

	AvgRT() float64
	MinRT() float64
	CurrentGoroutineNum() int32

	AddPassRequest(count uint64)
	AddRtAndCompleteRequest(rt, count uint64)
	AddBlockRequest(count uint64)
	AddErrorRequest(count uint64)
	IncreaseGoroutineNum()
	DecreaseGoroutineNum()

	Reset()
}
