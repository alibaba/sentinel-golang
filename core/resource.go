package core

// ResourceType represents classification of the resources
type ResourceType int32

const (
	ResTypeCommon ResourceType = iota
	ResTypeWeb
	ResTypeRPC
	ResTypeAPIGateway
	ResTypeDBSQL
	ResTypeCache
	ResTypeMQ
)

// TrafficType describes the traffic type: Inbound or OutBound
type TrafficType int32

const (
	// InBound represents the inbound traffic (e.g. provider)
	InBound TrafficType = iota
	// OutBound represents the outbound traffic (e.g. consumer)
	OutBound
)

// ResourceWrapper represents the invocation
type ResourceWrapper struct {
	// global unique resource name
	ResourceName string
	// resource classification
	Classification ResourceType
	// InBound or OutBound
	FlowType TrafficType
}
