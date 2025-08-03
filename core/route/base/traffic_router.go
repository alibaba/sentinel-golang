package base

type TrafficRouter struct {
	Host []string
	Http []*HTTPRoute
}

type Fallback struct {
	Host   string
	Subset string
}
