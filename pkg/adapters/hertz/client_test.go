package hertz

import (
	"context"
	"net/http"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func initSentinelForClient(t *testing.T) {
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}
	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "GET:/client_ping",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "/api/users/:id",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
		return
	}
}

func TestClientSentinelMiddleware(t *testing.T) {
	type args struct {
		opts    []ClientOption
		method  string
		reqPath string
	}
	type want struct {
		code  int
		isErr bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default get",
			args: args{
				opts:    []ClientOption{},
				method:  http.MethodGet,
				reqPath: "http://localhost:3000/client_ping",
			},
			want: want{
				code:  http.StatusOK,
				isErr: false,
			},
		},
		{
			name: "customize resource extract",
			args: args{
				opts: []ClientOption{
					WithClientResourceExtractor(func(ctx context.Context, req *protocol.Request, resp *protocol.Response) string {
						return "/api/users/:id"
					}),
				},
				method:  http.MethodGet,
				reqPath: "http://localhost:3000/api/users/123",
			},
			want: want{
				code:  http.StatusTooManyRequests,
				isErr: true,
			},
		},
		{
			name: "customize block fallback",
			args: args{
				opts: []ClientOption{
					WithClientBlockFallback(func(ctx context.Context, req *protocol.Request, resp *protocol.Response, blockError error) error {
						resp.SetStatusCode(http.StatusBadRequest)
						return blockError
					}),
				},
				method:  http.MethodGet,
				reqPath: "http://localhost:3000/client_ping",
			},
			want: want{
				code:  http.StatusBadRequest,
				isErr: true,
			},
		},
	}
	go func() {
		h := server.New(server.WithHostPorts(":3000"))
		h.GET("/client_ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.String(200, "pong")
		})
		h.GET("/api/users/:id", func(c context.Context, ctx *app.RequestContext) {
			ctx.String(200, "pong")
		})
		h.Spin()
	}()
	initSentinelForClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := client.NewClient()
			if err != nil {
				t.Fatalf("Unexpected error: %+v", err)
				return
			}
			c.Use(SentinelClientMiddleware(tt.args.opts...))
			req := &protocol.Request{}
			res := &protocol.Response{}
			req.SetMethod(tt.args.method)
			req.SetRequestURI(tt.args.reqPath)
			err = c.Do(context.Background(), req, res)
			assert.Equal(t, tt.want.isErr, err != nil)
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
}
