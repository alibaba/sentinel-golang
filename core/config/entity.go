package config

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

var props map[string]string

type Entity struct {
	// Version represents the format version of the entity.
	Version string

	Sentinel SentinelConfig
}

// SentinelConfig represent the general configuration of Sentinel.
type SentinelConfig struct {
	App struct {
		// Name represents the name of current running service.
		Name string
		// Type indicates the classification of the service (e.g. web service, API gateway).
		Type int32
	}
	// Log represents configuration items related to logging.
	Log LogConfig
	// Stat represents configuration items related to statistics.
	Stat StatConfig
}

// LogConfig represent the configuration of logging in Sentinel.
type LogConfig struct {
	// Dir represents the log directory path.
	Dir string
	// UsePid indicates whether the filename ends with the process ID (PID).
	UsePid bool `yaml:"usePid"`
	// Metric represents the configuration items of the metric log.
	Metric MetricLogConfig
}

// MetricLogConfig represents the configuration items of the metric log.
type MetricLogConfig struct {
	SingleFileMaxSize uint64 `yaml:"singleFileMaxSize"`
	MaxFileCount      uint32 `yaml:"maxFileCount"`
	FlushIntervalSec  uint32 `yaml:"flushIntervalSec"`
}

// StatConfig represents the configuration items of statistics.
type StatConfig struct {
	System SystemStatConfig `yaml:"system"`
}

// SystemStatConfig represents the configuration items of system statistics.
type SystemStatConfig struct {
	// CollectIntervalMs represents the collecting interval of the system metrics collector.
	CollectIntervalMs uint32 `yaml:"collectIntervalMs"`
}

func NewDefaultConfig() *Entity {
	return &Entity{
		Version: "v1",
		Sentinel: SentinelConfig{
			App: struct {
				Name string
				Type int32
			}{
				Name: UnknownProjectName,
				Type: DefaultAppType,
			},
			Log: LogConfig{
				Dir:    GetDefaultLogDir(),
				UsePid: false,
				Metric: MetricLogConfig{
					SingleFileMaxSize: DefaultMetricLogSingleFileMaxSize,
					MaxFileCount:      DefaultMetricLogMaxFileAmount,
					FlushIntervalSec:  DefaultMetricLogFlushIntervalSec,
				},
			},
			Stat: StatConfig{
				System: SystemStatConfig{
					CollectIntervalMs: DefaultSystemStatCollectIntervalMs,
				},
			},
		},
	}
}

func checkValid(conf *SentinelConfig) error {
	if conf == nil {
		return errors.New("Nil globalCfg")
	}
	if conf.App.Name == "" {
		return errors.New("App.Name is empty")
	}
	mc := conf.Log.Metric
	if mc.MaxFileCount <= 0 {
		return errors.New("Illegal metric log globalCfg: maxFileCount <= 0")
	}
	if mc.SingleFileMaxSize <= 0 {
		return errors.New("Illegal metric log globalCfg: singleFileMaxSize <= 0")
	}
	if conf.Stat.System.CollectIntervalMs == 0 {
		return errors.New("Bad system stat globalCfg: collectIntervalMs = 0")
	}
	return nil
}

func SetConfig(key, value string){
	if props == nil{
		props = make(map[string]string)
	}
	logger := logging.GetDefaultLogger()
	if key == "" || value == ""{
		logger.Error("Can't set config with empty key or value")
		return
	}
	props[key] = value
}

func GetConfig(key string)string{
	logger := logging.GetDefaultLogger()
	if key == ""{
		logger.Error("Can't get config with empty key")
		return ""
	}
	return props[key]
}
