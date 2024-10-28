package main

import (
	"flag"
	"log"
	"net"
	"time"

	api "github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var addressFlag = flag.String("server_address", ":8000", "Set the listen address for server")
var nodeCrashFlag = flag.Bool("node_crash", false, "Set the flag for whether to simulate node crash")

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"

func main() {
	flag.Parse()
	etcdReg, err := etcd.NewEtcdRegistry([]string{etcdAddr})
	addr, err := net.ResolveTCPAddr("tcp", *addressFlag)
	if err != nil {
		log.Fatal(err)
	}
	svr := api.NewServer(
		&HelloImpl{getIDWithAddress(*addressFlag), time.Now()},
		server.WithServiceAddr(addr),
		server.WithRegistry(etcdReg),
		server.WithServerBasicInfo(
			&rpcinfo.EndpointBasicInfo{
				ServiceName: serviceName,
			}),
	)
	err = svr.Run()
	if err != nil {
		log.Fatal(err)
	}
}
