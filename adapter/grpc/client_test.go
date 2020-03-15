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
	interceptor := SentinelUnaryClientIntercept()
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	opts ...grpc.CallOption) error {
		return errors.New(errMsgFake)
	}
	method := "/grpc.testing.TestService/UnaryCall"
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/UnaryCall",
				MetricType:      flow.QPS,
				Count:           1,
				ControlBehavior: flow.Reject,
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
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/UnaryCall",
				MetricType:      flow.QPS,
				Count:           0,
				ControlBehavior: flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, method, nil, nil, nil, invoker)
		assert.IsType(t, &base.BlockError{}, err)
	})
}

func TestStreamClientIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := SentinelStreamClientIntercept()
	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, errors.New(errMsgFake)
	}
	method := "/grpc.testing.TestService/StreamingOutputCall"
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/StreamingOutputCall",
				MetricType:      flow.QPS,
				Count:           1,
				ControlBehavior: flow.Reject,
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
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/StreamingOutputCall",
				MetricType:      flow.QPS,
				Count:           0,
				ControlBehavior: flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := interceptor(nil, nil, nil, method, streamer)
		assert.IsType(t, &base.BlockError{}, err)
		assert.Nil(t, rep)
	})
}
