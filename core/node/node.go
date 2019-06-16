package node

type Node interface {
	//TotalRequest() uint64
	//TotalPass() uint64
	TotalSuccess() uint64
	//BlockRequest() uint64
	//TotalError() uint64
	//PassQps() uint64
	//BlockQps() uint64
	//TotalQps() uint64
	//SuccessQps() uint64
	//MaxSuccessQps() uint64
	//ErrorQps() uint64
	//AvgRt() float32
	//MinRt() float32
	//CurGoroutineNum() uint64

	//PreviousBlockQps() uint64
	//PreviousPassQps() uint64
	//
	//Metrics() map[uint64]*metrics.MetricNode

	AddPassRequest(count uint32)
	//AddRtAndSuccess(rt uint64, success uint32)
	//
	//IncreaseBlockQps(count uint32)
	//IncreaseErrorQps(count uint32)
	//
	//IncreaseGoroutineNum()
	//DecreaseGoroutineNum()

	//Reset()
}
