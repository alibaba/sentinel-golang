package trpc

import (
	"context"
	"errors"
	"os"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/filter"
)

func TestMain(m *testing.M) {
	_ = sentinel.InitDefault()
	m.Run()
	os.Exit(0)
}

func TestSentinelServerFilter(t *testing.T) {
	const errMsgFake = "fake error"
	serverFilter := SentinelServerFilter()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}

	t.Run("success", func(t *testing.T) {
		ctx, msg := codec.WithNewMessage(context.Background())
		msg.WithCalleeServiceName("trpc.test.helloworld.Greeter")
		msg.WithServerRPCName("/trpc.test.helloworld.Greeter/SayHello")

		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "trpc.test.helloworld.Greeter:/trpc.test.helloworld.Greeter/SayHello",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := serverFilter(ctx, nil, handler)
		assert.EqualError(t, err, errMsgFake)
		assert.Nil(t, rep)
		// Test for recording the biz error.
		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode("trpc.test.helloworld.Greeter:/trpc.test.helloworld.Greeter/SayHello").GetQPS(base.MetricEventError)))

		t.Run("second fail", func(t *testing.T) {
			rep, err := serverFilter(ctx, nil, handler)
			assert.IsType(t, &base.BlockError{}, err)
			assert.Nil(t, rep)
		})
	})

	successHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}
	t.Run("fail", func(t *testing.T) {
		ctx, msg := codec.WithNewMessage(context.Background())
		msg.WithCalleeServiceName("trpc.test.fail.Service")
		msg.WithServerRPCName("/trpc.test.fail.Service/Method")

		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "trpc.test.fail.Service:/trpc.test.fail.Service/Method",
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := serverFilter(ctx, nil, successHandler)
		assert.IsType(t, &base.BlockError{}, err)
		assert.Nil(t, rep)
	})
}

func TestSentinelServerFilterWithBlockFallback(t *testing.T) {
	const fallbackMsg = "fallback"
	bf := func(ctx context.Context, req interface{}, blockErr *base.BlockError) (interface{}, error) {
		return fallbackMsg, nil
	}
	serverFilter := SentinelServerFilter(WithBlockFallback(bf))

	ctx, msg := codec.WithNewMessage(context.Background())
	msg.WithCalleeServiceName("trpc.test.fallback.Service")
	msg.WithServerRPCName("/trpc.test.fallback.Service/Method")

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	var _, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "trpc.test.fallback.Service:/trpc.test.fallback.Service/Method",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})
	assert.Nil(t, err)
	rep, err := serverFilter(ctx, nil, handler)
	assert.Nil(t, err)
	assert.Equal(t, fallbackMsg, rep)
}

func TestSentinelServerFilterWithResourceExtract(t *testing.T) {
	serverFilter := SentinelServerFilter(
		WithResourceExtract(func(ctx context.Context, req interface{}) string {
			return "custom-resource"
		}),
	)

	ctx, _ := codec.WithNewMessage(context.Background())
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	var _, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "custom-resource",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})
	assert.Nil(t, err)
	rep, err := serverFilter(ctx, nil, handler)
	assert.IsType(t, &base.BlockError{}, err)
	assert.Nil(t, rep)
}

func TestSentinelServerFilter_FilterRegistration(t *testing.T) {
	serverFilter := SentinelServerFilter()
	clientFilter := SentinelClientFilter()

	filter.Register("sentinel-test", serverFilter, clientFilter)

	registeredServer := filter.GetServer("sentinel-test")
	registeredClient := filter.GetClient("sentinel-test")

	assert.NotNil(t, registeredServer)
	assert.NotNil(t, registeredClient)
}
