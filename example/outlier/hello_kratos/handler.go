package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/go-kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type server struct {
	address    string
	networkErr bool
	isCrash    bool
	done       chan struct{}
	pb.UnimplementedGreeterServer
}

func NewServer() (transport.Server, transport.Server) {
	// 初始化 http server
	httpSrv := http.NewServer(
		http.Address(*httpAddressFlag),
		http.Middleware(
			recovery.Recovery(),
		),
	)

	// 初始化 grpc server
	grpcSrv := grpc.NewServer(
		grpc.Address(*grpcAddressFlag),
		grpc.Middleware(
			recovery.Recovery(),
		),
	)

	// 在服务器上注册服务
	s := &server{address: *grpcAddressFlag}
	pb.RegisterGreeterServer(grpcSrv, s)
	pb.RegisterGreeterHTTPServer(httpSrv, s)
	go func() {
		start := 10 * time.Second
		end := 15 * time.Second
		timer1 := time.NewTimer(start)
		timer2 := time.NewTimer(end)

		<-timer1.C
		s.isCrash = true

		<-timer2.C
		s.isCrash = false
		s.done <- struct{}{}
	}()
	return grpcSrv, httpSrv
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (resp *pb.HelloReply, err error) {
	if s.isCrash {
		if s.networkErr { // 如果是网络故障
			<-s.done
			return resp, nil
		}
		// 如果是服务故障
		return &pb.HelloReply{Message: fmt.Sprintf("Welcome %s,I am %s!", in.Name, s.address)}, fmt.Errorf("server error")
	}
	resp = &pb.HelloReply{Message: fmt.Sprintf("Welcome %s,I am %s!", in.Name, s.address)}
	return resp, nil
}
