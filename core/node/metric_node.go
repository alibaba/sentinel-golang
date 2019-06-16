package node

type MetricNode struct {
	Timestamp  uint64
	PassQps    uint64
	BlockQps   uint64
	SuccessQps uint64
	ErrorQps   uint64
	Rt         uint64
	Resource   string
}
