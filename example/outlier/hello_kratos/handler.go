package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/go-kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type server struct {
	pb.UnimplementedGreeterServer
	id        int
	startTime time.Time
}

// NewServer inits http server and grpc server
func NewServer() (transport.Server, transport.Server) {
	httpSrv := http.NewServer(
		http.Address(*httpAddressFlag),
		http.Middleware(recovery.Recovery()),
	)
	grpcSrv := grpc.NewServer(
		grpc.Address(*grpcAddressFlag),
		grpc.Middleware(recovery.Recovery()),
	)
	s := &server{id: getIDWithAddress(*grpcAddressFlag), startTime: time.Now()}
	pb.RegisterGreeterServer(grpcSrv, s)
	pb.RegisterGreeterHTTPServer(httpSrv, s)
	return grpcSrv, httpSrv
}

func getIDWithAddress(address string) int {
	return int(address[len(address)-1] - '0')
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (resp *pb.HelloReply, err error) {
	message := fmt.Sprintf("Welcome %s,I am node%d", in.Name, s.id)
	if *nodeCrashFlag {
		return &pb.HelloReply{Message: message}, nil
	}
	faultStartTime := s.startTime.Add(5 * time.Second).Add(time.Duration(s.id) * 5 * time.Second)
	faultEndTime := faultStartTime.Add(20 * time.Second)
	currentTime := time.Now()
	// If currentTime is in the time range of the business error
	if currentTime.After(faultStartTime) && currentTime.Before(faultEndTime) {
		return nil, errors.New("internal server error")
	}
	return &pb.HelloReply{Message: message}, nil
}
