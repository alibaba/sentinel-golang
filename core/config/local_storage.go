package config

const (
	UnknownProjectName = "unknown_go_service"

	LogDirEnvKey     = "SENTINEL_LOG_DIR"
	LogNamePidEnvKey = "SENTINEL_LOG_USE_PID"

	AppNameEnvKey = "SENTINEL_APP_NAME"
	AppTypeEnvKey = "SENTINEL_APP_TYPE"

	DefaultAppType                    int32  = 0
	DefaultMetricLogFlushIntervalSec  uint32 = 1
	DefaultMetricLogSingleFileMaxSize uint64 = 1024 * 1024 * 50
	DefaultMetricLogMaxFileAmount     uint32 = 8
)

func AppName() string {
	// TODO
	return "ahas-go-service"
}

func AppType() int32 {
	return DefaultAppType
}

func MetricLogFlushIntervalSec() uint32 {
	return DefaultMetricLogFlushIntervalSec
}

func MetricLogSingleFileMaxSize() uint64 {
	return DefaultMetricLogSingleFileMaxSize
}

func MetricLogMaxFileAmount() uint32 {
	return DefaultMetricLogMaxFileAmount
}
