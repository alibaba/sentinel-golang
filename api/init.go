package api

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/constant"
	"github.com/alibaba/sentinel-golang/core/log/metric"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/util"
	"os"
)

// InitDefault initializes Sentinel using the configuration from system
// environment and the default value.
func InitDefault() error {
	return initSentinel("")
}

// Init loads Sentinel general configuration from the given YAML file
// and initializes Sentinel. Note that the logging module will be initialized
// using the configuration from system environment or the default value.
func Init(configPath string) error {
	return initSentinel(configPath)
}

//
// The priority: ENV > yaml file > default configuration
func initSentinel(configPath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	//Firstly, get config file path
	if util.IsBlank(configPath) {
		// If the config file path is absent, Sentinel will try to resolve it from the system env.
		configPath = os.Getenv(constant.ConfFilePathEnvKey)
	}
	if util.IsBlank(configPath) {
		configPath = constant.DefaultConfigFilename
	}
	// load config from yaml file
	// if use don't set config path, then use default config
	err = config.LoadFromYamlFile(configPath)
	if err != nil {
		return err
	}
	// Secondly, use variable from ENV to override config
	err = config.OverrideFromSystemEnv()
	if err != nil {
		return err
	}

	err = config.InitializeLogConfig(config.LogBaseDir(), config.LogUsePid())
	if err != nil {
		return err
	}

	initCoreComponents()
	return err
}

func initCoreComponents()  {
	metric.InitTask()
	system.InitCollector(config.SystemStatCollectIntervalMs())
}
