package micro

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/stretchr/testify/assert"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/outlier"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/util"

	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

func initClient(t *testing.T) client.Client {
	r := etcd.NewRegistry(
		registry.Addrs("localhost:2379"),
	)
	s := selector.NewSelector(
		selector.Registry(r),
		selector.SetStrategy(selector.RoundRobin),
	)
	c := client.NewClient(
		// set the selector
		client.Selector(s),
		// add the client wrapper
		client.Wrap(NewClientWrapper(
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
	return c
}

func TestClientLimiter1(t *testing.T) {
	c := initClient(t)
	req := c.NewRequest("helloworld", "Helloworld.Call", &proto.Request{}, client.WithContentType("application/json"))
	rsp := &proto.Response{}
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               req.Method(),
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)

		err = c.Call(context.TODO(), req, rsp)
		assert.Nil(t, err)
		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))

		t.Run("second fail", func(t *testing.T) {
			err := c.Call(context.TODO(), req, rsp)
			assert.EqualError(t, err, FakeErrorMsg)
			assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))
		})
	})
}

func TestClientLimiter2(t *testing.T) {
	c := initClient(t)
	req := c.NewRequest("helloworld", "Helloworld.Call", &proto.Request{}, client.WithContentType("application/json"))
	rsp := &proto.Response{}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
			{
				Rule: &circuitbreaker.Rule{
					Resource:         req.Method(),
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
