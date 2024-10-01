package base

type TrafficContext struct {
	Path        string
	Uri         string
	ServiceName string
	Group       string
	Version     string
	MethodName  string
	ParamTypes  []string
	Args        []interface{}
	Headers     map[string]string
	Baggage     map[string]string
}
