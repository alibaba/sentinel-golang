package main

import (
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/etcdv3"
	"github.com/alibaba/sentinel-golang/util"
	"log"
	"math/rand"
	"time"
)

func main() {
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}
	config.SetConfig(etcdv3.EndPoints,"127.0.0.1:2379")
	handler := datasource.NewSinglePropertyHandler(flow.FlowRulesConvert, flow.FlowRulesUpdate)
	client := etcdv3.NewEtcdDataSource("flow",handler)
	if client == nil {
		log.Fatal("Create etcd client failed")
		return
	}

	ch := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}
