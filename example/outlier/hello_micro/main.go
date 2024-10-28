package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"

	pb "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"
const version = "latest"

func main() {
	etcdReg := etcd.NewRegistry(registry.Addrs(etcdAddr))
	srv := micro.NewService()
	srv.Init(
		micro.Name(serviceName),
		micro.Version(version),
		micro.Registry(etcdReg),
	)
	if err := pb.RegisterTestHandler(srv.Server(), &TestHandler{
		getIDWithAddress(srv.Server().Options().Address),
		time.Now(),
	}); err != nil {
		logger.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
