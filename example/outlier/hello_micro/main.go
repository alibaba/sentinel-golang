package main

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"

	pb "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

var (
	service = "helloworld"
	version = "latest"
)

func main() {
	etcdReg := etcd.NewRegistry(
		registry.Addrs("localhost:2379"),
	)

	// Create service
	srv := micro.NewService()
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Registry(etcdReg),
	)

	// Register handler
	if err := pb.RegisterTestHandler(srv.Server(), &TestHandler{}); err != nil {
		logger.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
