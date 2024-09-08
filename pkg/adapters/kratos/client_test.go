package kratos

import (
	"context"
	"testing"
	"time"

	pb "github.com/go-kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/outlier"
)

const serviceName = "example.helloworld"
const etcdAddr = "127.0.0.1:2379"

func initOutlierClient(t *testing.T) pb.GreeterClient {
	// new discovery with etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdAddr},
	})
	if err != nil {
		t.Fatal(err)
	}
	etcdReg := etcd.New(client)

	// Due to the limitations of the gRPC framework, selector can
	// only be injected using a global balancer.
	selector.SetGlobalSelector(wrr.NewBuilder())

	endpoint := "discovery:///" + serviceName
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(etcdReg),
		grpc.WithNodeFilter(OutlierClientFilter),
		grpc.WithMiddleware(OutlierClientMiddleware),
	)
	if err != nil {
		t.Fatal(err)
	}
	return pb.NewGreeterClient(conn)
}

func TestOutlierClientMiddleware(t *testing.T) {
	c := initOutlierClient(t)
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("success", func(t *testing.T) {
		var _, err = outlier.LoadRules([]*outlier.Rule{
			{
				Rule: &circuitbreaker.Rule{
					Resource:         serviceName,
					Strategy:         circuitbreaker.ErrorCount,
					RetryTimeoutMs:   3000,
					MinRequestAmount: 1,
					StatIntervalMs:   1000,
					Threshold:        1.0,
				},
				EnableActiveRecovery: true,
				MaxEjectionPercent:   1.0,
				RecoveryInterval:     2000,
				MaxRecoveryAttempts:  5,
			},
		})
		assert.Nil(t, err)
		passCount, testCount := 0, 200
		req := &pb.HelloRequest{Name: "Bob"}
		for i := 0; i < testCount; i++ {
			ctx := metadata.NewClientContext(context.Background(), metadata.New())
			rsp, err := c.SayHello(ctx, req)
			t.Log(rsp, err)
			if err == nil {
				passCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		t.Logf("Results: %d out of %d requests were successful\n", passCount, testCount)
	})
}
