package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	api "github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var addressFlag = flag.String("server_address", ":8888", "Set the listen address for server")
var errorFlag = flag.Bool("network_error", true, "Set the error type for server")

func main() {
	flag.Parse()
	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Fatal(err)
	}

	addr, _ := net.ResolveTCPAddr("tcp", *addressFlag)
	fmt.Println(*addressFlag, *errorFlag)
	svr := api.NewServer(
		NewHello(),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithServerBasicInfo(
			&rpcinfo.EndpointBasicInfo{
				ServiceName: "example.hello",
			}),
	)

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
