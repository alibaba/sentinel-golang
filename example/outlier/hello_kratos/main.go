package main

import (
	"flag"
	"log"

	etcdregitry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	etcdclient "go.etcd.io/etcd/client/v3"
)

var httpAddressFlag = flag.String("http_server_address", ":8888", "Set the listen address for http server")
var grpcAddressFlag = flag.String("grpc_server_address", ":9999", "Set the listen address for grpc server")

func main() {
	flag.Parse()
	client, err := etcdclient.New(etcdclient.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个 registry 对象，就是对 ectd client 操作的一个包装
	r := etcdregitry.New(client)
	grpcSrv, httpSrv := NewServer()
	app := kratos.New(
		kratos.Name("helloworld"), // 服务名称
		kratos.Server(
			httpSrv,
			grpcSrv,
		),
		kratos.Registrar(r), // 填入etcd连接(etcd作为服务中心)
	)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
