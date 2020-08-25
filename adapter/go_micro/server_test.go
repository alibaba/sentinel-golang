package go_micro

import (
	"context"
	"github.com/alibaba/sentinel-golang/adapter/go_micro/proto"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/micro/go-micro/v2"
	"log"
	"sync"
	"testing"
	"time"
)

type TestHandler struct {}
func (h *TestHandler) Ping(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
	rsp.Result = "Pong"
	return nil
}

func TestServerLimiter(t *testing.T) {

	server := micro.NewService(
		micro.Name("sentinel.test.server"),
		micro.Version("latest"),
		micro.WrapHandler(NewHandlerWrapper()),
	)

	_ = proto.RegisterTestHandler(server.Server(), &TestHandler{})

	go server.Run()

	time.Sleep(time.Second)

	c := server.Client()
	req := c.NewRequest("sentinel.test.server", "Test.Ping", &proto.Request{})

	err := sentinel.InitDefault()
	if err != nil {
		log.Fatal(err)
	}

	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "Test.Ping",
			MetricType:      flow.QPS,
			Count:           LimitCount,
			ControlBehavior: flow.Reject,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	var rsp = &proto.Response{}

	wg := new(sync.WaitGroup)
	wg.Add(30)
	for i := 0; i < LimitCount * 3 ; i++ {
		go func() {
			err := c.Call(context.TODO(), req, rsp)
			if err != nil {
				t.Logf("Got err when call, %v", err)
			} else  {
				t.Log("Simulate call finished")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}