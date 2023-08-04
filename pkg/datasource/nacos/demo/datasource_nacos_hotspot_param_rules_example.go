package main

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
)

func main() {
	startHotSpotRuleModule()

	// nacos server info
	sc := []constant.ServerConfig{
		{
			ContextPath: "/nacos",
			Port:        8848,
			IpAddr:      "127.0.0.1",
		},
	}
	// nacos client info
	cc := constant.ClientConfig{
		TimeoutMs: 5000,
	}
	//build nacos config client
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		fmt.Printf("Fail to create client, err: %+v", err)
		return
	}
	h := datasource.NewHotSpotParamRulesHandler(datasource.HotSpotParamRuleJsonArrayParser)
	//sentinel-go is nacos configuration management Group in flow control
	//flow is nacos configuration management DataId in flow control
	nds, err := nacos.NewNacosDataSource(client, "sentinel-go", "flow", h)
	if err != nil {
		fmt.Printf("Fail to create nacos data source client, err: %+v", err)
		return
	}
	// initialize *NacosDataSource and load rule
	err = nds.Initialize()
	if err != nil {
		fmt.Printf("Fail to initialize nacos data source client, err: %+v", err)
		return
	}

	// Simulation of the request
	ch := make(chan struct{})
	<-ch
}

func startHotSpotRuleModule() {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	_, err = hotspot.LoadRules([]*hotspot.Rule{
		{
			Resource:        "abc",
			MetricType:      hotspot.QPS,
			ControlBehavior: hotspot.Reject,
			ParamIndex:      1,
			Threshold:       100,
			BurstCount:      0,
			DurationInSec:   1000,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	logging.Info("[HotSpot Reject] Sentinel Go hot-spot param flow control demo is running. You may see the pass/block metric in the metric log.")
	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("abc", sentinel.WithArgs(true, rand.Uint32()%30, "sentinel"))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					fmt.Println("block...")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println("pass...")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
}
