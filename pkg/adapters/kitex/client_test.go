package kitex

import (
	"context"
	"errors"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/stretchr/testify/assert"
)

const FakeErrorMsg = "fake error for testing"

func TestSentinelClientMiddleware(t *testing.T) {
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	c, err := hello.NewClient("hello",
		client.WithMiddleware(SentinelClientMiddleware(
			WithBlockFallback(bf))))
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
