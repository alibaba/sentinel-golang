package nacos_v2

const (
	DefaultTimeoutMs uint64 = 4000
)

type Config struct {
	TimeoutMs uint64 `yaml:"timeoutMs"`
}
