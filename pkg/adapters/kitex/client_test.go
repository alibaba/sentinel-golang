package kitex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
	"github.com/stretchr/testify/assert"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/outlier"
)

const FakeErrorMsg = "fake error for testing"

func TestSentinelClientMiddleware(t *testing.T) {
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	c, err := hello.NewClient("hello",
		client.WithMiddleware(SentinelClientMiddleware(WithBlockFallback(bf))))
	if err != nil {
		t.Fatal(err)
	}
	err = sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}
	req := &api.Request{}
	t.Run("success", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "hello:echo",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		_, err = c.Echo(context.Background(), req)
		assert.NotNil(t, err)
		assert.NotEqual(t, FakeErrorMsg, err.Error())
		t.Run("second fail", func(t *testing.T) {
			_, err = c.Echo(context.Background(), req)
			assert.NotNil(t, err)
			assert.Equal(t, FakeErrorMsg, err.Error())
		})
	})
}

func initClient(t *testing.T) hello.Client {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		panic(err)
	}
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	c, err := hello.NewClient("example.hello",
		client.WithResolver(OutlierClientResolver(r)),
		client.WithMiddleware(OutlierClientMiddleware(WithBlockFallback(bf))),
	)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func TestSentinelClientMiddleware1(t *testing.T) {
	c := initClient(t)
	req := &api.Request{Message: "Bob"}
	// callopt.WithHostPort("localhost:8888")
	resp, err := c.Echo(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
	time.Sleep(time.Second)

	err = sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "hello:echo",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		_, err = c.Echo(context.Background(), req)
		assert.NotNil(t, err)
		assert.NotEqual(t, FakeErrorMsg, err.Error())
		t.Run("second fail", func(t *testing.T) {
			_, err = c.Echo(context.Background(), req)
			assert.NotNil(t, err)
			assert.Equal(t, FakeErrorMsg, err.Error())
		})
	})
}

func TestOutlierClient(t *testing.T) {
	c := initClient(t)
	req := &api.Request{Message: "Bob"}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
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
				RecoveryInterval:     2000,
				MaxRecoveryAttempts:  5,
			},
		})
		assert.Nil(t, err)
		passCount := 0
		testCount := 100
		for i := 0; i < testCount; i++ {
			resp, err := c.Echo(context.Background(), req)
			fmt.Println(resp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Printf("pass %f%%\n", float64(passCount)*100/float64(testCount))
	})
}
