package resources

type XdsRouteConfig struct {
	Name         string
	VirtualHosts map[string]XdsVirtualHost
}

type XdsVirtualHost struct {
	Name    string
	Domains []string
	Routes  []XdsRoute
}

type XdsRoute struct {
	Name   string
	Match  XdsRouteMatch
	Action XdsRouteAction
}

type XdsRouteAction struct {
	Cluster        string
	ClusterWeights []XdsClusterWeight
}

type XdsClusterWeight struct {
	Name   string
	Weight uint32
}

type XdsRouteMatch interface {
	MatchPath(path string) bool
	// default use headers to match the meta.
	MatchMeta(map[string]string) bool
}

type HTTPRouteMatch struct {
	Path          string
	Prefix        string
	Regex         string
	CaseSensitive bool
	Headers       Matchers
}

func (rm *HTTPRouteMatch) MatchPath(path string) bool {
	if rm.Path != "" {
		return rm.Path == path
	}
	// default prefix
	return rm.Prefix == "/"
}

func (rm *HTTPRouteMatch) MatchMeta(md map[string]string) bool {
	return rm.Headers.Match(md)
}
