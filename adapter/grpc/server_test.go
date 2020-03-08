package grpc

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"google.golang.org/grpc/status"
	testpb "google.golang.org/grpc/test/grpc_testing"
)

func testServerInitSentinel(t *testing.T) {
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}

	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "/grpc.testing.TestService/UnaryCall",
			MetricType:      flow.QPS,
			Count:           10,
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

// func TestStreamServerIntercept(t *testing.T) {
// 	testServerInitSentinel(t)
// 	sopts := []grpc.ServerOption{
// 		grpc.StreamInterceptor(SentinelStreamServerIntercept()),
// 	}
// 	ss := &stubServer{
// 		fullDuplexCall: func(stream testpb.TestService_FullDuplexCallServer) error {
// 			return stream.SendMsg(&testpb.SimpleRequest{ })
// 		},
// 	}
// 	if err := ss.Start(sopts); err != nil {
// 		t.Fatalf("Error starting endpoint server: %v", err)
// 	}
// 	defer ss.Stop()
//
// 	resp, err := ss.client.StreamingInputCall(context.Background())
// 	if s, ok := status.FromError(err); !ok || s.Code() != codes.OK {
// 		t.Fatalf("ss.client.UnaryCall(context.Background(), _) = %v, %v; want nil, <status with Code()=OK>", resp, err)
// 	}
//
// 	var recv testpb.SimpleRequest
// 	assert.Equal(t, nil, resp.RecvMsg(&recv))
// }

func TestUnaryServerIntercept(t *testing.T) {
	testServerInitSentinel(t)
	sopts := []grpc.ServerOption{
		grpc.UnaryInterceptor(SentinelUnaryServerIntercept()),
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
	if err := ss.Start(sopts); err != nil {
		t.Fatalf("Error starting endpoint server: %v", err)
	}
	defer ss.Stop()

	resp, err := ss.client.UnaryCall(context.Background(), &testpb.SimpleRequest{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.OK {
		t.Fatalf("ss.client.UnaryCall(context.Background(), _) = %v, %v; want nil, <status with Code()=OK>", resp, err)
	}

	respBytes := resp.GetPayload().GetBody()
	if string(respBytes) != "unary call" {
		t.Fatalf("invalid response: want=%s, but got=%s", "unary call", resp)
	}
}

// stubServer is a server that is easy to customize within individual test
// cases.
type stubServer struct {
	// Guarantees we satisfy this interface; panics if unimplemented methods are called.
	testpb.TestServiceServer

	// Customizable implementations of server handlers.
	emptyCall      func(ctx context.Context, in *testpb.Empty) (*testpb.Empty, error)
	unaryCall      func(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error)
	fullDuplexCall func(stream testpb.TestService_FullDuplexCallServer) error

	// A client connected to this service the test may use.  Created in Start().
	client testpb.TestServiceClient
	cc     *grpc.ClientConn
	s      *grpc.Server

	addr string // address of listener

	cleanups []func() // Lambdas executed in Stop(); populated by Start().

	r *manual.Resolver
}

func (ss *stubServer) EmptyCall(ctx context.Context, in *testpb.Empty) (*testpb.Empty, error) {
	return ss.emptyCall(ctx, in)
}

func (ss *stubServer) UnaryCall(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error) {
	return ss.unaryCall(ctx, in)
}

func (ss *stubServer) FullDuplexCall(stream testpb.TestService_FullDuplexCallServer) error {
	return ss.fullDuplexCall(stream)
}

// Start starts the server and creates a client connected to it.
func (ss *stubServer) Start(sopts []grpc.ServerOption, dopts ...grpc.DialOption) error {
	r, cleanup := manual.GenerateAndRegisterManualResolver()
	ss.r = r
	ss.cleanups = append(ss.cleanups, cleanup)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf(`net.Listen("tcp", "localhost:0") = %v`, err)
	}
	ss.addr = lis.Addr().String()
	ss.cleanups = append(ss.cleanups, func() { lis.Close() })

	s := grpc.NewServer(sopts...)
	testpb.RegisterTestServiceServer(s, ss)
	go s.Serve(lis)
	ss.cleanups = append(ss.cleanups, s.Stop)
	ss.s = s

	target := ss.r.Scheme() + ":///" + ss.addr

	opts := append([]grpc.DialOption{grpc.WithInsecure()}, dopts...)
	cc, err := grpc.Dial(target, opts...)
	if err != nil {
		return fmt.Errorf("grpc.Dial(%q) = %v", target, err)
	}
	ss.cc = cc
	ss.r.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: ss.addr}}})
	if err := ss.waitForReady(cc); err != nil {
		return err
	}

	ss.cleanups = append(ss.cleanups, func() { cc.Close() })

	ss.client = testpb.NewTestServiceClient(cc)
	return nil
}

func (ss *stubServer) waitForReady(cc *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		s := cc.GetState()
		if s == connectivity.Ready {
			return nil
		}
		if !cc.WaitForStateChange(ctx, s) {
			// ctx got timeout or canceled.
			return ctx.Err()
		}
	}
}

func (ss *stubServer) Stop() {
	for i := len(ss.cleanups) - 1; i >= 0; i-- {
		ss.cleanups[i]()
	}
}

func newPayload(t testpb.PayloadType, size int32) (*testpb.Payload, error) {
	if size < 0 {
		return nil, fmt.Errorf("requested a response with invalid length %d", size)
	}
	body := make([]byte, size)
	switch t {
	case testpb.PayloadType_COMPRESSABLE:
	case testpb.PayloadType_UNCOMPRESSABLE:
		return nil, fmt.Errorf("PayloadType UNCOMPRESSABLE is not supported")
	default:
		return nil, fmt.Errorf("unsupported payload type: %d", t)
	}
	return &testpb.Payload{
		Type: t,
		Body: body,
	}, nil
}
