package base

type Headers struct {
	Request  *HeaderOperations
	Response *HeaderOperations
}

type HeaderOperations struct {
	Set    map[string]string
	Add    map[string]string
	Remove []string
}
