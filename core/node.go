package core

// Node holds real-time statistics for resources.
type Node interface {
	// total  = pass + blocked
	TotalCountInMinute() uint64
	PassCountInMinute() uint64
	BlockCountInMinute() uint64
	CompleteCountInMinute() uint64
	ErrorCountInMinute() uint64

	TotalQPS() float64
	PassQPS() float64
	BlockQPS() float64
	CompleteQPS() float64
	ErrorQPS() float64

	AvgRT() float64
	MinRT() float64
	CurrentGoroutineNum() uint32

	AddPassRequest(count uint64)
	AddRtAndCompleteRequest(rt, count uint64)
	AddBlockRequest(count uint64)
	AddErrorRequest(count uint64)
	IncreaseGoroutineNum()
	DecreaseGoroutineNum()

	Reset()
}
