package micro

import (
	"context"
	"testing"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/stretchr/testify/assert"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"
const version = "latest"

func initOutlierClient(t *testing.T) client.Client {
	etcdReg := etcd.NewRegistry(registry.Addrs(etcdAddr))
	sel := selector.NewSelector(
		selector.Registry(etcdReg),
		selector.SetStrategy(selector.RoundRobin),
	)
	srv := micro.NewService(
		micro.Name(serviceName),
		micro.Version(version),
		micro.Selector(sel),
		micro.WrapClient(NewOutlierClientWrapper()),
	)
	return srv.Client()
}

func TestOutlierClientMiddleware(t *testing.T) {
	c := initOutlierClient(t)
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
			{
				Rule: &circuitbreaker.Rule{
					Resource:         serviceName,
					Strategy:         circuitbreaker.ErrorCount,
					RetryTimeoutMs:   3000,
					MinRequestAmount: 1,
					StatIntervalMs:   1000,
					Threshold:        1.0,
				},
				EnableActiveRecovery: true,
				MaxEjectionPercent:   1.0,
				RecoveryInterval:     2000,
				MaxRecoveryAttempts:  5,
			},
		})
		assert.Nil(t, err)
		passCount, testCount := 0, 200
		req := c.NewRequest(serviceName, "Test.Ping", &proto.Request{},
			client.WithContentType("application/json"))
		for i := 0; i < testCount; i++ {
			rsp := &proto.Response{}
			err = c.Call(context.Background(), req, rsp)
			t.Log(rsp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		t.Logf("Results: %d out of %d requests were successful\n", passCount, testCount)
	})
}
