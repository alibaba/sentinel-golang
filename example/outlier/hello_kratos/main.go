package main

import (
	"flag"
	"log"

	etcdregitry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	etcdclient "go.etcd.io/etcd/client/v3"
)

var httpAddressFlag = flag.String("http_server_address", ":8000", "Set the listen address for http server")
var grpcAddressFlag = flag.String("grpc_server_address", ":9000", "Set the listen address for grpc server")
var nodeCrashFlag = flag.Bool("node_crash", false, "Set the flag for whether to simulate node crash")

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"

func main() {
	flag.Parse()
	client, err := etcdclient.New(etcdclient.Config{
		Endpoints: []string{etcdAddr},
	})
	if err != nil {
		log.Fatal(err)
	}
	etcdReg := etcdregitry.New(client)
	grpcSrv, httpSrv := NewServer()
	app := kratos.New(
		kratos.Name(serviceName),
		kratos.Server(
			httpSrv,
			grpcSrv,
		),
		kratos.Registrar(etcdReg),
	)
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
