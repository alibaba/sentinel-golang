package base

const (
	TotalInBoundResourceName = "__total_inbound_traffic__"

	DefaultMaxResourceAmount uint32 = 10000

	// default 10*1000/500 = 20
	DefaultSampleCount uint32 = 20
	// default 10s
	DefaultIntervalInMs uint32 = 10000

	DefaultStatisticMaxRt = int64(5000)
)
