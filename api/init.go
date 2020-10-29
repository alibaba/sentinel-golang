package api

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/log/metric"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

// environment and the default value.
func InitDefault() error {
	return initSentinel("")
}

// InitWithConfig initializes Sentinel using given config.
func InitWithConfig(confEntity *config.Entity) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	err = config.CheckValid(confEntity)
	if err != nil {
		return err
	}
	config.SetDefaultConfig(confEntity)
	if err = config.OverrideConfigFromEnvAndInitLog(); err != nil {
		return err
	}
	return initCoreComponents()
}

// Init loads Sentinel general configuration from the given YAML file
// and initializes Sentinel.
func InitWithConfigFile(configPath string) error {
	return initSentinel(configPath)
}

// initCoreComponents init core components with default config
// it's better SetDefaultConfig before initCoreComponents
func initCoreComponents() error {
	if err := metric.InitTask(); err != nil {
		return err
	}
	if !util.IsWindowsOS() {
		system.InitCollector(config.SystemStatCollectIntervalMs())
	} else {
		logging.Warn("[Init initCoreComponents] system metric collect is not available for system module in windows")
	}

	if config.UseCacheTime() {
		util.StartTimeTicker()
	}

	return nil
}

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
	// Initialize general config and logging module.
	if err = config.InitConfig(configPath); err != nil {
		return err
	}
	return initCoreComponents()
}
