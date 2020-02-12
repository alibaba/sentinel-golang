package config

import "github.com/pkg/errors"

type Config struct {
	Version  string
	Sentinel struct {
		App struct {
			// Name represents the name of current running service.
			Name string
			// Type indicates the classification of the service (e.g. web service, API gateway).
			Type int32
		}
		Log struct {
			// Dir represents the log directory path.
			Dir string
			// UsePid indicates whether the filename ends with the process ID (PID).
			UsePid bool
			// Metric represents the configuration items of the metric log.
			Metric struct {
				SingleFileMaxSize uint64
				MaxFileCount      uint32
				FlushIntervalSec  uint32
			}
		}
	}
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
}

type Entity struct {
	// Version represents the format version of the entity.
	Version string

	Sentinel SentinelConfig
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
			Log: LogConfig{Metric: MetricLogConfig{SingleFileMaxSize: DefaultMetricLogSingleFileMaxSize, MaxFileCount: DefaultMetricLogMaxFileAmount, FlushIntervalSec: DefaultMetricLogFlushIntervalSec}},
		},
	}
}

func checkValid(conf *SentinelConfig) error {
	if conf == nil {
		return errors.New("nil config")
	}
	if conf.App.Name == "" {
		return errors.New("app.name cannot be empty")
	}
	mc := conf.Log.Metric
	if mc.MaxFileCount <= 0 {
		return errors.New("Bad metric log config: maxFileCount <= 0")
	}
	if mc.SingleFileMaxSize <= 0 {
		return errors.New("Bad metric log config: singleFileMaxSize <= 0")
	}
	return nil
}
