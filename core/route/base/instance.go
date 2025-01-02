package base

type Instance struct {
	AppName        string
	Host           string
	Port           int
	Metadata       map[string]string
	TargetInstance interface{}
}
