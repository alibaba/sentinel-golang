package trpc

import (
	"context"
	"errors"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
	"trpc.group/trpc-go/trpc-go/codec"
)

func TestSentinelClientFilter(t *testing.T) {
	const errMsgFake = "fake error"
	clientFilter := SentinelClientFilter()
	handler := func(ctx context.Context, req, rsp interface{}) error {
		return errors.New(errMsgFake)
	}

	t.Run("success", func(t *testing.T) {
		ctx, msg := codec.WithNewMessage(context.Background())
		msg.WithCalleeServiceName("trpc.test.client.Service")
		msg.WithClientRPCName("/trpc.test.client.Service/Method")

		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "trpc.test.client.Service:/trpc.test.client.Service/Method",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = clientFilter(ctx, nil, nil, handler)
		assert.EqualError(t, err, errMsgFake)
		// Test for recording the biz error.
		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode("trpc.test.client.Service:/trpc.test.client.Service/Method").GetQPS(base.MetricEventError)))

		t.Run("second fail", func(t *testing.T) {
			err := clientFilter(ctx, nil, nil, handler)
			assert.IsType(t, &base.BlockError{}, err)
		})
	})

	successHandler := func(ctx context.Context, req, rsp interface{}) error {
		return nil
	}
	t.Run("fail", func(t *testing.T) {
		ctx, msg := codec.WithNewMessage(context.Background())
		msg.WithCalleeServiceName("trpc.test.client.fail.Service")
		msg.WithClientRPCName("/trpc.test.client.fail.Service/Method")

		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "trpc.test.client.fail.Service:/trpc.test.client.fail.Service/Method",
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = clientFilter(ctx, nil, nil, successHandler)
		assert.IsType(t, &base.BlockError{}, err)
	})
}

func TestSentinelClientFilterWithResourceExtract(t *testing.T) {
	clientFilter := SentinelClientFilter(
		WithResourceExtract(func(ctx context.Context, req interface{}) string {
			return "custom-client-resource"
		}),
	)

	ctx, _ := codec.WithNewMessage(context.Background())
	handler := func(ctx context.Context, req, rsp interface{}) error {
		return nil
	}

	var _, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "custom-client-resource",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})
	assert.Nil(t, err)
	err = clientFilter(ctx, nil, nil, handler)
	assert.IsType(t, &base.BlockError{}, err)
}
