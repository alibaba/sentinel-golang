package kratos

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/go-kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
)

func initClient(t *testing.T) pb.GreeterClient {
	// new etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		panic(err)
	}
	// new discovery with etcd client
	dis := etcd.New(client)

	endpoint := "discovery:///example.helloworld"
	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
	selector.SetGlobalSelector(wrr.NewBuilder())

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(dis),
		grpc.WithNodeFilter(OutlierClientFilter),
		grpc.WithMiddleware(OutlierClientMiddleware),
	)
	if err != nil {
		panic(err)
	}
	client2 := pb.NewGreeterClient(conn)
	return client2
}

func TestClient(t *testing.T) {
	client2 := initClient(t)
	req := &pb.HelloRequest{Name: "World"}
	reply, err := client2.SayHello(context.Background(), req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Infof("Greeting: %s", reply.GetMessage())
}

func TestClientLimiter2(t *testing.T) {
	c := initClient(t)
	req := &pb.HelloRequest{Name: "Bob"}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
			{
				Rule: &circuitbreaker.Rule{
					Resource:         "my_rpc_service",
					Strategy:         circuitbreaker.ErrorCount,
					RetryTimeoutMs:   3000,
					MinRequestAmount: 1,
					StatIntervalMs:   1000,
					Threshold:        1.0,
				},
				EnableActiveRecovery: false,
				MaxEjectionPercent:   1,
				RecoveryInterval:     2000,
				MaxRecoveryAttempts:  5,
			},
		})
		assert.Nil(t, err)
		passCount := 0
		testCount := 100
		for i := 0; i < testCount; i++ {
			ctx := metadata.NewClientContext(context.TODO(), metadata.New())
			rsp, err := c.SayHello(ctx, req)
			fmt.Println(rsp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Printf("pass %f%%\n", float64(passCount)*100/float64(testCount))
	})
}
