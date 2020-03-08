package api

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/log/metric"
	"github.com/alibaba/sentinel-golang/core/system"
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
	err = config.InitConfig(configPath)
	initCoreComponents()
	return err
}

func initCoreComponents()  {
	metric.InitTask()
	system.InitCollector(config.SystemStatCollectIntervalMs())
}
