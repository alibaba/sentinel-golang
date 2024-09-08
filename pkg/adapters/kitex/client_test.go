package kitex

import (
	"context"
	"errors"
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

func initOutlierClient(t *testing.T) hello.Client {
	resolver, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := hello.NewClient("example.helloworld",
		client.WithResolver(OutlierClientResolver(resolver)),
		client.WithMiddleware(OutlierClientMiddleware()),
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
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
		passCount, testCount := 0, 200
		req := &api.Request{Message: "Bob"}
		for i := 0; i < testCount; i++ {
			resp, err := c.Echo(context.Background(), req)
			t.Log(resp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		t.Logf("Results: %d out of %d requests were successful\n", passCount, testCount)
	})
}
