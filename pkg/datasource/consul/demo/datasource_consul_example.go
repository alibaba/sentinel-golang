package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/pkg/datasource/consul"
	"github.com/hashicorp/consul/api"
)

func main() {
	ch := make(chan struct{})
	startFlowModule()
	ds := startConsulDs()
	defer ds.Close()
	<-ch
}

func startFlowModule() {
	// We should initialize Sentinel first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}

	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "some-test",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              10,
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
		return
	}

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
}

func startConsulDs() datasource.DataSource {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
	})
	if err != nil {
		fmt.Println("Failed to instance consul client")
		os.Exit(1)
	}
	// Note: need to put key "example-consul-cb-rules" in consul server.
	ds, err := consul.NewDataSource("example-consul-cb-rules",
		// customize consul client
		consul.WithConsulClient(client),
		// preset property handlers
		consul.WithPropertyHandlers(datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser)),
	)

	if err != nil {
		fmt.Println("Failed to instance consul datasource")
		os.Exit(1)
	}

	if err := ds.Initialize(); err != nil {
		fmt.Println("Failed to initialize consul datasource")
		os.Exit(1)
	}

	return ds
}
