package grpc

import (
	"context"
	"testing"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	testpb "google.golang.org/grpc/test/grpc_testing"
)

func testClientInitSentinel(t *testing.T) {
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}

	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "/grpc.testing.TestService/UnaryCall",
			MetricType:      flow.QPS,
			Count:           1,
			ControlBehavior: flow.Reject,
		},
		{
			Resource:        "/grpc.testing.TestService/StreamingInputCall",
			MetricType:      flow.QPS,
			Count:           1,
			ControlBehavior: flow.Reject,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
		return
	}
}

func TestUnaryClientIntercept(t *testing.T) {
	testClientInitSentinel(t)
	dopts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(SentinelUnaryClientIntercept()),
	}
	ss := &stubServer{
		unaryCall: func(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error) {
			payload, err := newPayload(testpb.PayloadType_COMPRESSABLE, 0)
			if err != nil {
				return nil, status.Errorf(codes.Aborted, "failed to make payload: %v", err)
			}

			payload.Body = []byte("unary call")

			return &testpb.SimpleResponse{
				Payload: payload,
			}, nil
		},
	}
	if err := ss.Start([]grpc.ServerOption{}, dopts...); err != nil {
		t.Fatalf("Error starting endpoint server: %v", err)
	}
	defer ss.Stop()

	// success on first calling
	{
		resp, err := ss.client.UnaryCall(context.Background(), &testpb.SimpleRequest{})
		s, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.OK, s.Code())

		respBytes := resp.GetPayload().GetBody()
		assert.Equal(t, "unary call", string(respBytes))
	}

	// fail on second calling when calling interval < 1 second
	{
		_, err := ss.client.UnaryCall(context.Background(), &testpb.SimpleRequest{})
		s, ok := status.FromError(err)
		assert.Equal(t, false, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		berr, ok := err.(*base.BlockError)
		assert.Equal(t, true, ok)
		assert.Equal(t, "SentinelBlockException: Flow", berr.Error())
	}

	// success on third calling which in new leap window
	{
		time.Sleep(time.Second)
		resp, err := ss.client.UnaryCall(context.Background(), &testpb.SimpleRequest{})
		s, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.OK, s.Code())

		respBytes := resp.GetPayload().GetBody()
		assert.Equal(t, "unary call", string(respBytes))
	}
}
