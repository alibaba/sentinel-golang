package constant

// configuration
const (
	UnknownProjectName = "unknown_go_service"

	ConfFilePathEnvKey = "SENTINEL_CONFIG_FILE_PATH"
	AppNameEnvKey      = "SENTINEL_APP_NAME"
	AppTypeEnvKey      = "SENTINEL_APP_TYPE"
	LogDirEnvKey       = "SENTINEL_LOG_DIR"
	LogNamePidEnvKey   = "SENTINEL_LOG_USE_PID"

	DefaultConfigFilename       = "sentinel.yml"
	DefaultAppType        int32 = 0

	DefaultMetricLogFlushIntervalSec  uint32 = 1
	DefaultMetricLogSingleFileMaxSize uint64 = 1024 * 1024 * 50
	DefaultMetricLogMaxFileAmount     uint32 = 8
	DefaultSystemStatCollectIntervalMs uint32 = 1000
)

// global variable
const (
	TotalInBoundResourceName = "__total_inbound_traffic__"

	DefaultMaxResourceAmount uint32 = 10000

	DefaultSampleCount uint32 = 2
	DefaultIntervalMs  uint32 = 1000

	// default 10*1000/500 = 20
	DefaultSampleCountTotal uint32 = 20
	// default 10s (total length)
	DefaultIntervalMsTotal uint32 = 10000

	DefaultStatisticMaxRt = int64(60000)
)