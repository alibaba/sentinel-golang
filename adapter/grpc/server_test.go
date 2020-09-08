package grpc

import (
	"context"
	"errors"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	_ = sentinel.InitDefault()
	m.Run()
}

func TestStreamServerIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := NewStreamServerInterceptor()
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return errors.New(errMsgFake)
	}
	info := &grpc.StreamServerInfo{
		FullMethod: "/grpc.testing.TestService/StreamingInputCall",
	}

	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/StreamingInputCall",
				MetricType:             flow.QPS,
				Count:                  1,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
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
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/StreamingInputCall",
				MetricType:             flow.QPS,
				Count:                  0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		err = interceptor(nil, nil, info, handler)
		assert.IsType(t, &base.BlockError{}, err)
	})
}

func TestUnaryServerIntercept(t *testing.T) {
	const errMsgFake = "fake error"
	interceptor := NewUnaryServerInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New(errMsgFake)
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				MetricType:             flow.QPS,
				Count:                  1,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := interceptor(nil, nil, info, handler)
		assert.EqualError(t, err, errMsgFake)
		assert.Nil(t, rep)
		// Test for recording the biz error.
		assert.EqualValues(t, 1, int(stat.GetResourceNode(info.FullMethod).GetQPS(base.MetricEventError)))

		t.Run("second fail", func(t *testing.T) {
			rep, err := interceptor(nil, nil, info, handler)
			assert.IsType(t, &base.BlockError{}, err)
			assert.Nil(t, rep)

			assert.EqualValues(t, 1, int(stat.GetResourceNode(info.FullMethod).GetQPS(base.MetricEventError)))
		})
	})

	successHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "abc", nil
	}
	t.Run("fail", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "/grpc.testing.TestService/UnaryCall",
				MetricType:             flow.QPS,
				Count:                  0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		rep, err := interceptor(nil, nil, info, successHandler)
		assert.IsType(t, &base.BlockError{}, err)
		assert.Nil(t, rep)
	})
}
