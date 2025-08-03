package base

import "fmt"

type HTTPRouteDestination struct {
	Weight      int
	Destination *Destination
	Headers     Headers // TODO modifies headers
}

func (H HTTPRouteDestination) String() string {
	return fmt.Sprintf("{Weight: %v, Destination: %+v}\n", H.Weight, H.Destination)
}

type Destination struct {
	Host     string
	Subset   string
	Port     uint32
	Fallback *HTTPRouteDestination
}
