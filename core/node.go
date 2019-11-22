package core

type node interface {
	// total  = pass + blocked
	TotalCountInMinute() uint64
	PassCountInMinute() uint64
	BlockCountInMinute() uint64
	CompleteCountInMinute() uint64
	ExceptionCountInMinute() uint64

	TotalQps() uint64
	PassQps() uint64
	BlockQps() uint64
	CompleteQps() uint64
	ExceptionQps() uint64

	AvgRt() uint64
	CurrentGoroutineNum() uint64

	AddPassRequest(count uint64)
	AddRtAndCompleteRequest(rt, count uint64)
	AddBlockRequest(count uint64)
	AddExceptionRequest(count uint64)
	IncreaseGoroutineNum()
	DecreaseGoroutineNum()

	Reset()
}
