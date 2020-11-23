package benchmark

import (
	"log"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
)

func InitSentinel() {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0
	conf.Sentinel.Stat.System.CollectIntervalMs = 0
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}
