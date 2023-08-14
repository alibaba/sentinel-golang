package base

type VirtualWorkload struct {
	Host          string
	trafficPolicy *TrafficPolicy
	Subsets       []*Subset
}

type Subset struct {
	Name   string
	Labels map[string]string
}
