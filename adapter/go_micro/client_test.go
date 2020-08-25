package go_micro

import (
	"context"
	"github.com/alibaba/sentinel-golang/adapter/go_micro/proto"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry/memory"
	"log"
	"sync"
	"testing"
)

const LimitCount = 10

func TestClientLimiter(t *testing.T) {
	// setup
	r := memory.NewRegistry()
	s := selector.NewSelector(selector.Registry(r))

	c := client.NewClient(
		// set the selector
		client.Selector(s),
		// add the breaker wrapper
		client.Wrap(NewClientWrapper()),
	)

	req := c.NewRequest("sentinel.test.server", "Test.Ping", &proto.Request{}, client.WithContentType("application/json"))

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

	rsp := &proto.Response{}
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
