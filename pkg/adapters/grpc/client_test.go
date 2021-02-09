package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryClientIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := NewUnaryClientInterceptor(WithUnaryClientResourceExtractor(
		func(ctx context.Context, method string, i interface{}, conn *grpc.ClientConn) string {
			return "client:" + method
		}))
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		opts ...grpc.CallOption) error {
		return errors.New(errMsgFake)
	}
	method := "/grpc.testing.TestService/UnaryCall"
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "client:" + method,
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, method, nil, nil, nil, invoker)
		assert.EqualError(t, err, errMsgFake)
		t.Run("second fail", func(t *testing.T) {
			err = interceptor(nil, method, nil, nil, nil, invoker)
			assert.IsType(t, &base.BlockError{}, err)
		})
	})

	t.Run("fail", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "client:" + method,
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, method, nil, nil, nil, invoker)
		assert.IsType(t, &base.BlockError{}, err)
	})
}

func TestStreamClientIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := NewStreamClientInterceptor(WithStreamClientResourceExtractor(
		func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string) string {
			return "client:" + method
		}))
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, errors.New(errMsgFake)
	}
	method := "/grpc.testing.TestService/StreamingOutputCall"
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "client:/grpc.testing.TestService/StreamingOutputCall",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := interceptor(nil, nil, nil, method, streamer)
		assert.EqualError(t, err, errMsgFake)
		assert.Nil(t, rep)
		t.Run("second fail", func(t *testing.T) {
			rep, err := interceptor(nil, nil, nil, method, streamer)
			assert.IsType(t, &base.BlockError{}, err)
			assert.Nil(t, rep)
		})
	})

	t.Run("fail", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "client:/grpc.testing.TestService/StreamingOutputCall",
				Threshold:              0.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := interceptor(nil, nil, nil, method, streamer)
		assert.IsType(t, &base.BlockError{}, err)
		assert.Nil(t, rep)
	})
}
