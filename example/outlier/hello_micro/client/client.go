package main

import (
	"context"
	"log"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
	microAdapter "github.com/alibaba/sentinel-golang/pkg/adapters/micro"
	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"
const version = "latest"

func initOutlierClient() client.Client {
	etcdReg := etcd.NewRegistry(registry.Addrs(etcdAddr))
	sel := selector.NewSelector(
		selector.Registry(etcdReg),
		selector.SetStrategy(selector.RoundRobin),
	)
	srv := micro.NewService(
		micro.Name(serviceName),
		micro.Version(version),
		micro.Selector(sel),
		micro.WrapClient(microAdapter.NewClientWrapper(
			microAdapter.WithEnableOutlier(func(ctx context.Context) bool {
				return true
			}))),
	)
	return srv.Client()
}

func main() {
	c := initOutlierClient()
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatal(err)
	}
	_, err = outlier.LoadRules([]*outlier.Rule{
		{
			Rule: &circuitbreaker.Rule{
				Resource:         serviceName,
				Strategy:         circuitbreaker.ErrorCount,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 1,
				StatIntervalMs:   1000,
				Threshold:        1.0,
			},
			EnableActiveRecovery: false,
			MaxEjectionPercent:   1.0,
			RecoveryIntervalMs:   2000,
			MaxRecoveryAttempts:  5,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	passCount, testCount := 0, 200
	req := c.NewRequest(serviceName, "Test.Ping", &proto.Request{},
		client.WithContentType("application/json"))
	for i := 0; i < testCount; i++ {
		rsp := &proto.Response{}
		err = c.Call(context.Background(), req, rsp)
		log.Println(rsp, err)
		if err == nil {
			passCount++
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("Results: %d out of %d requests were successful\n", passCount, testCount)
}
