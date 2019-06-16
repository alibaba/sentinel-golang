package base

const (
	INBOUND = iota
	OUTBOUND
)

type ResourceWrapper struct {
	// unique resource name
	ResourceName string
	//
	ResourceType int
}

type SlotResultStatus int8

const (
	ResultStatusOk = iota
	ResultStatusBlocked
)

type SlotResult struct {
	Status        SlotResultStatus
	BlockedReason string
}
