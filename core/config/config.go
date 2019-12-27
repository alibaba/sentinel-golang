package config

var configMap = make(map[string]string, 0)

const (
	UnknownAppName = "unknown_go_service"

	LogDirEnvKey     = "CSP_SENTINEL_LOG_DIR"
	LogNamePidEnvKey = "CSP_SENTINEL_LOG_USE_PID"

	AppNameEnvKey = "CSP_SENTINEL_APP_NAME"
	AppTypeEnvKey = "CSP_SENTINEL_APP_TYPE"
)

func AppName() string {
	return ""
}
