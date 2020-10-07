package main

import (
	"math/rand"
	"os"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/logging"
)

func main() {
	cfg := config.NewDefaultConfig()
	cfg.Sentinel.Log.Logger = logging.NewConsoleLogger()
	cfg.Sentinel.Log.Metric.FlushIntervalSec = 0
	cfg.Sentinel.Stat.System.CollectIntervalMs = 0
	err := sentinel.InitWithConfig(cfg)
	if err != nil {
		logging.Error(err, "fail")
		os.Exit(1)
	}
	logging.SetGlobalLoggerLevel(logging.DebugLevel)
	ch := make(chan struct{})

	r1 := &isolation.Rule{
		Resource:   "abc",
		MetricType: isolation.Concurrency,
		Threshold:  12,
	}
	_, err = isolation.LoadRules([]*isolation.Rule{r1})
	if err != nil {
		logging.Error(err, "fail")
		os.Exit(1)
	}

	for i := 0; i < 15; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithBatchCount(1))
				if b != nil {
					logging.Info("blocked", "reason", b.BlockType().String(), "rule", b.TriggeredRule(), "snapshot", b.TriggeredValue())
					time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
				} else {
					logging.Info("passed")
					time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
					e.Exit()
				}
			}
		}()
	}
	<-ch
}
