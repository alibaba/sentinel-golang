package kitex

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/stretchr/testify/assert"
)

// HelloImpl implements the last service interface defined in the IDL.
type HelloImpl struct{}

// Echo implements the HelloImpl interface.
func (s *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{Message: req.Message}
	return
}

func TestSentinelServerMiddleware(t *testing.T) {
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	srv := hello.NewServer(new(HelloImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "hello"}),
		server.WithMiddleware(SentinelServerMiddleware(
			WithBlockFallback(bf),
		)))
	go srv.Run()
	defer srv.Stop()
	time.Sleep(1 * time.Second)

	c, err := hello.NewClient("hello", client.WithHostPorts(":8888"))
	assert.Nil(t, err)

	err = sentinel.InitDefault()
	assert.Nil(t, err)
	req := &api.Request{}
	t.Run("success", func(t *testing.T) {
		_, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "hello:echo",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		_, err := c.Echo(context.TODO(), req)
		assert.Nil(t, err)

		t.Run("second fail", func(t *testing.T) {
			_, err := c.Echo(context.TODO(), req)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), FakeErrorMsg))
		})
	})
}
