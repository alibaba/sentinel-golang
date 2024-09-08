package micro

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/stretchr/testify/assert"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

func initClient(t *testing.T) client.Client {
	etcdReg := etcd.NewRegistry(
		registry.Addrs("localhost:2379"),
	)
	s := selector.NewSelector(
		selector.Registry(etcdReg),
		selector.SetStrategy(selector.RoundRobin),
	)
	srv := micro.NewService(
		micro.Name("helloworld"),
		micro.Version("latest"),
		micro.Registry(etcdReg),
		micro.Selector(s),
		micro.WrapClient(NewOutlierClientWrapper(
			// add custom fallback function to return a fake error for assertion
			WithClientBlockFallback(
				func(ctx context.Context, request client.Request, blockError *base.BlockError) error {
					return errors.New(FakeErrorMsg)
				}),
		)),
	)
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}
	return srv.Client()
}

func TestOutlierClient(t *testing.T) {
	c := initClient(t)
	req := c.NewRequest("helloworld", "Test.Ping", &proto.Request{}, client.WithContentType("application/json"))
	rsp := &proto.Response{}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
			{
				Rule: &circuitbreaker.Rule{
					Resource:         req.Service(),
					Strategy:         circuitbreaker.ErrorCount,
					RetryTimeoutMs:   3000,
					MinRequestAmount: 1,
					StatIntervalMs:   1000,
					Threshold:        1.0,
				},
				EnableActiveRecovery: false,
				MaxEjectionPercent:   1,
				RecoveryInterval:     2000,
				MaxRecoveryAttempts:  5,
			},
		})
		assert.Nil(t, err)
		passCount := 0
		testCount := 100
		for i := 0; i < testCount; i++ {
			err = c.Call(context.TODO(), req, rsp)
			fmt.Println(rsp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Printf("pass %f%%\n", float64(passCount)*100/float64(testCount))
	})
}
