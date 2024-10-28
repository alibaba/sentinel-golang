package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
	"github.com/alibaba/sentinel-golang/pkg/adapters/kitex"
)

func initOutlierClient() hello.Client {
	resolver, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Fatal(err)
	}
	c, err := hello.NewClient("example.helloworld",
		client.WithResolver(kitex.OutlierClientResolver(resolver)),
		client.WithMiddleware(kitex.SentinelClientMiddleware(
			kitex.WithEnableOutlier(func(ctx context.Context) bool {
				return true
			}))),
	)
	if err != nil {
		log.Fatal(err)
	}
	return c
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
				Resource:         "example.helloworld",
				Strategy:         circuitbreaker.ErrorCount,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 1,
				StatIntervalMs:   1000,
				Threshold:        1.0,
			},
			EnableActiveRecovery: true,
			MaxEjectionPercent:   1.0,
			RecoveryIntervalMs:   2000,
			MaxRecoveryAttempts:  5,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	passCount, testCount := 0, 200
	req := &api.Request{Message: "Bob"}
	for i := 0; i < testCount; i++ {
		resp, err := c.Echo(context.Background(), req)
		log.Println(resp, err)
		if err == nil {
			passCount++
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("Results: %d out of %d requests were successful\n", passCount, testCount)
}
