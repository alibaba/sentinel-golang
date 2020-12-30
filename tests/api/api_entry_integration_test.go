package api

import (
	"log"
	"runtime/debug"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func initSentinel() {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	conf.Sentinel.Log.Metric.FlushIntervalSec = 0
	conf.Sentinel.Stat.System.CollectIntervalMs = 0
	conf.Sentinel.Stat.System.CollectMemoryIntervalMs = 0
	conf.Sentinel.Stat.System.CollectCpuIntervalMs = 0
	conf.Sentinel.Stat.System.CollectLoadIntervalMs = 0
	err := api.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}

func TestAdaptiveFlowControl(t *testing.T) {
	initSentinel()
	util.SetClock(util.NewMockClock())

	rs := "hello0"
	rule := flow.Rule{
		Resource:               rs,
		TokenCalculateStrategy: flow.MemoryAdaptive,
		ControlBehavior:        flow.Reject,
		StatIntervalInMs:       1000,
		LowMemUsageThreshold:   5,
		HighMemUsageThreshold:  1,
		MemLowWaterMarkBytes:   1 * 1024,
		MemHighWaterMarkBytes:  2 * 1024,
	}
	rule1 := rule
	ok, err := flow.LoadRules([]*flow.Rule{&rule1})
	assert.True(t, ok)
	assert.Nil(t, err)

	// mock memory usage < MemLowWaterMarkBytes, QPS threshold is 2
	system_metric.SetSystemMemoryUsage(512)
	for i := 0; i < 5; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		entry.Exit()
	}
	_, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}

	// clear statistic
	util.Sleep(time.Second * 2)
	// QPS threshold is 3
	system_metric.SetSystemMemoryUsage(1536)
	for i := 0; i < 3; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		entry.Exit()
	}
	_, blockError = api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}

	// clear statistic
	util.Sleep(time.Second * 2)
	t.Log("start to test memory based adaptive flow control")
	// QPS threshold is 3
	system_metric.SetSystemMemoryUsage(2049)
	for i := 0; i < 1; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		entry.Exit()
	}
	_, blockError = api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}
}

func TestAdaptiveFlowControl2(t *testing.T) {
	debug.SetGCPercent(-1)
	initSentinel()
	rs := "hello0"
	rule := flow.Rule{
		Resource:               rs,
		TokenCalculateStrategy: flow.MemoryAdaptive,
		ControlBehavior:        flow.Reject,
		StatIntervalInMs:       1000,
		LowMemUsageThreshold:   150,
		HighMemUsageThreshold:  10,
		MemLowWaterMarkBytes:   100998840320,
		MemHighWaterMarkBytes:  268435456000,
	}
	ok, err := flow.LoadRules([]*flow.Rule{&rule})
	assert.True(t, ok)
	assert.Nil(t, err)
	system_metric.SetSystemMemoryUsage(136794800128)
	_, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.Nil(t, blockError)
}
