package config

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

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
	// UseCacheTime indicates whether to cache time(ms)
	UseCacheTime bool `yaml:"useCacheTime"`
}

// LogConfig represent the configuration of logging in Sentinel.
type LogConfig struct {
	// logger indicates that using logger to replace default logging.
	Logger logging.Logger
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

// NewDefaultConfig creates a new default config entity.
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
				Logger: nil,
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
			UseCacheTime: true,
		},
	}
}

func CheckValid(entity *Entity) error {
	if entity == nil {
		return errors.New("Nil entity")
	}
	if len(entity.Version) == 0 {
		return errors.New("Empty version")
	}
	return checkConfValid(&entity.Sentinel)
}

func checkConfValid(conf *SentinelConfig) error {
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

func (entity *Entity) AppName() string {
	return entity.Sentinel.App.Name
}

func (entity *Entity) AppType() int32 {
	return entity.Sentinel.App.Type
}

func (entity *Entity) LogBaseDir() string {
	return entity.Sentinel.Log.Dir
}

func (entity *Entity) Logger() logging.Logger {
	return entity.Sentinel.Log.Logger
}

// LogUsePid returns whether the log file name contains the PID suffix.
func (entity *Entity) LogUsePid() bool {
	return entity.Sentinel.Log.UsePid
}

func (entity *Entity) MetricLogFlushIntervalSec() uint32 {
	return entity.Sentinel.Log.Metric.FlushIntervalSec
}

func (entity *Entity) MetricLogSingleFileMaxSize() uint64 {
	return entity.Sentinel.Log.Metric.SingleFileMaxSize
}

func (entity *Entity) MetricLogMaxFileAmount() uint32 {
	return entity.Sentinel.Log.Metric.MaxFileCount
}

func (entity *Entity) SystemStatCollectIntervalMs() uint32 {
	return entity.Sentinel.Stat.System.CollectIntervalMs
}

func (entity *Entity) UseCacheTime() bool {
	return entity.Sentinel.UseCacheTime
}
