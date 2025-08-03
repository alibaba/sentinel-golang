package base

type HTTPRoute struct {
	Name  string
	Match []*HTTPMatchRequest
	Route []*HTTPRouteDestination
}

func (h *HTTPRoute) IsMatch(context *TrafficContext) bool {
	for _, match := range h.Match {
		if match.IsMatch(context) {
			return true
		}
	}
	return false
}
