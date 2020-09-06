package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/nacos"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

type Counter struct {
	pass  *int64
	block *int64
	total *int64
}

func main() {
	counter := Counter{pass: new(int64), block: new(int64), total: new(int64)}

	if err := sentinel.InitDefault(); err != nil {
		fmt.Println(err)
		return
	}

	// For testing
	if err := logging.ResetGlobalLogger(logging.NewConsoleLogger("nacos-datasource-example")); err != nil {
		fmt.Println(err)
		return
	}

	//nacos server info
	sc := []constant.ServerConfig{
		{
			ContextPath: "/nacos",
			Port:        8848,
			IpAddr:      "127.0.0.1",
		},
	}
	//nacos client info
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
	h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
	//sentinel-go is nacos configuration management Group in flow control
	//flow is nacos configuration management DataId in flow control
	nds, err := nacos.NewNacosDataSource(client, "sentinel-go", "flow", h)
	if err != nil {
		fmt.Printf("Fail to create nacos data source client, err: %+v", err)
		return
	}
	//initialize *NacosDataSource and load rule
	err = nds.Initialize()
	if err != nil {
		fmt.Printf("Fail to initialize nacos data source client, err: %+v", err)
		return
	}
	//Starting counter
	go timerTask(&counter)

	//Simulation of the request
	ch := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for {
				atomic.AddInt64(counter.total, 1)
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					atomic.AddInt64(counter.block, 1)
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					atomic.AddInt64(counter.pass, 1)
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}

//statistic print
func timerTask(counter *Counter) {
	fmt.Println("begin to statistic!!!")
	var (
		oldTotal, oldPass, oldBlock int64
	)
	for {
		time.Sleep(1 * time.Second)
		globalTotal := atomic.LoadInt64(counter.total)
		oneSecondTotal := globalTotal - oldTotal
		oldTotal = globalTotal

		globalPass := atomic.LoadInt64(counter.pass)
		oneSecondPass := globalPass - oldPass
		oldPass = globalPass

		globalBlock := atomic.LoadInt64(counter.block)
		oneSecondBlock := globalBlock - oldBlock
		oldBlock = globalBlock
		fmt.Println(util.CurrentTimeMillis()/1000, "total:", oneSecondTotal, " pass:", oneSecondPass, " block:", oneSecondBlock)
	}
}
