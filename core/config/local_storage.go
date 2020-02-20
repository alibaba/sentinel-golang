package config

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	UnknownProjectName = "unknown_go_service"

	ConfFilePathEnvKey = "SENTINEL_CONFIG_FILE_PATH"

	AppNameEnvKey = "SENTINEL_APP_NAME"
	AppTypeEnvKey = "SENTINEL_APP_TYPE"

	DefaultConfigFilename                    = "sentinel.yml"
	DefaultAppType                    int32  = 0
	DefaultMetricLogFlushIntervalSec  uint32 = 1
	DefaultMetricLogSingleFileMaxSize uint64 = 1024 * 1024 * 50
	DefaultMetricLogMaxFileAmount     uint32 = 8
)

var localConf = NewDefaultConfig()

func InitConfig() error {
	return InitConfigFromFile("")
}

func InitConfigFromFile(filePath string) error {
	if util.IsBlank(filePath) {
		// If the file path is absent, Sentinel will try to resolve it from the system env.
		filePath = os.Getenv(ConfFilePathEnvKey)
		if util.IsBlank(filePath) {
			filePath = DefaultConfigFilename
		}
	}
	err := loadFromYamlFile(filePath)
	if err != nil {
		return err
	}
	loadFromSystemEnv()

	logger := logging.GetDefaultLogger()
	logger.Infof("App name resolved: %s", AppName())

	return nil
}

func loadFromYamlFile(filePath string) error {
	if filePath == DefaultConfigFilename {
		if _, err := os.Stat(DefaultConfigFilename); err != nil {
			return nil
		}
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &localConf)
	if err != nil {
		return err
	}
	return nil
}

func loadFromSystemEnv() {
	if appName := os.Getenv(AppNameEnvKey); !util.IsBlank(appName) {
		localConf.Sentinel.App.Name = appName
	}
	appTypeStr := os.Getenv(AppTypeEnvKey)
	appType, err := strconv.ParseInt(appTypeStr, 10, 32)
	if err != nil {

	} else {
		localConf.Sentinel.App.Type = int32(appType)
	}
}

func AppName() string {
	return localConf.Sentinel.App.Name
}

func AppType() int32 {
	return localConf.Sentinel.App.Type
}

func MetricLogFlushIntervalSec() uint32 {
	return localConf.Sentinel.Log.Metric.FlushIntervalSec
}

func MetricLogSingleFileMaxSize() uint64 {
	return localConf.Sentinel.Log.Metric.SingleFileMaxSize
}

func MetricLogMaxFileAmount() uint32 {
	return localConf.Sentinel.Log.Metric.MaxFileCount
}
