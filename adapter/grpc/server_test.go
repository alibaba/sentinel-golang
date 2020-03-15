package grpc

import (
	"context"
	"errors"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	_ = sentinel.InitDefault()
	m.Run()
}

func TestStreamServerIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := SentinelStreamServerIntercept()
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "/grpc.testing.TestService/StreamingInputCall",
	}

	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/StreamingInputCall",
				MetricType:      flow.QPS,
				Count:           1,
				ControlBehavior: flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, nil, info, handler)
		assert.EqualError(t, err, errMsgFake)
		t.Run("second fail", func(t *testing.T) {
			err = interceptor(nil, nil, info, handler)
			assert.IsType(t, &base.BlockError{}, err)
		})
	})

	t.Run("fail", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.FlowRule{
			{
				Resource:        "/grpc.testing.TestService/StreamingInputCall",
				MetricType:      flow.QPS,
				Count:           0,
				ControlBehavior: flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, nil, info, handler)
		assert.IsType(t, &base.BlockError{}, err)
	})
}

func TestUnaryServerIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := SentinelUnaryServerIntercept()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}
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
		rep, err := interceptor(nil, nil, info, handler)
		assert.EqualError(t, err, errMsgFake)
		assert.Nil(t, rep)
		t.Run("second fail", func(t *testing.T) {
			rep, err := interceptor(nil, nil, info, handler)
			assert.IsType(t, &base.BlockError{}, err)
			assert.Nil(t, rep)
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
		rep, err := interceptor(nil, nil, info, handler)
		assert.IsType(t, &base.BlockError{}, err)
		assert.Nil(t, rep)
	})
}

