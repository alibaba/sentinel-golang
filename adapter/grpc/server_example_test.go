package grpc

import (
	"context"
	"net"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
)

type greeter struct {
	Message string
}

func (g *greeter) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "hello"}, nil
}

func exampleInitSentinel() {
	// 1.0 init sentinel
	err := sentinel.InitDefault()
	if err != nil {
		panic(err)
	}

	// 2.0 load rules
	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "/grpc.testing.TestService/UnaryCall",
			MetricType:      flow.QPS,
			Count:           5,
			ControlBehavior: flow.Reject,
		},
		{
			Resource:        "/grpc.testing.TestService/StreamCall",
			MetricType:      flow.QPS,
			Count:           8,
			ControlBehavior: flow.Reject,
		},
	})
	if err != nil {
		panic(err)
	}

}

func ExampleServerIntercept() {
	// 1.0 init sentinel
	exampleInitSentinel()
	// 2.0 start grpc server, and inject interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(SentinelUnaryServerIntercept()),
		grpc.StreamInterceptor(SentinelStreamServerIntercept()),
	)
	helloworld.RegisterGreeterServer(s, new(greeter))

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	_ = s.Serve(lis)
}

func ExampleServerIntercept_CustomizeResourceName() {
	// 1.0 init sentinel
	exampleInitSentinel()
	// 2.0 customize resouce name injector
	extractUnaryResource := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) string {
		if peer, ok := peer.FromContext(ctx); ok {
			// set client addr as resource name
			return peer.Addr.String()
		}
		return info.FullMethod
	}

	extractStreamResource := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo) string {
		if peer, ok := peer.FromContext(ss.Context()); ok {
			// set client addr as resource name
			return peer.Addr.String()
		}
		return info.FullMethod
	}

	// 3.0 start grpc server, and inject interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			SentinelUnaryServerIntercept(
				WithUnaryServerResourceExtractor(extractUnaryResource),
			),
		),
		// todo(gorexlv): add resource extractor for stream
		grpc.StreamInterceptor(
			SentinelStreamServerIntercept(
				WithStreamServerResourceExtractor(extractStreamResource),
			),
		),
	)
	helloworld.RegisterGreeterServer(s, new(greeter))

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	_ = s.Serve(lis)
}

func ExampleServerIntercept_CustomizeErrorHandler() {
	// 1.0 init sentinel
	exampleInitSentinel()
	// 2.0 customize resource name injector
	fallbackUnaryIntercept := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	be *base.BlockError) (interface{}, error) {
		return nil, nil
	}
	fallbackStreamIntercept := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
	be *base.BlockError) error {
		return nil
	}
	// 2.0 start grpc server, and inject interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			SentinelUnaryServerIntercept(
				WithUnaryServerBlockFallback(fallbackUnaryIntercept),
			),
		),
		grpc.StreamInterceptor(
			SentinelStreamServerIntercept(
				WithStreamServerBlockFallback(fallbackStreamIntercept),
			),
		),
	)
	helloworld.RegisterGreeterServer(s, new(greeter))

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	_ = s.Serve(lis)
}
