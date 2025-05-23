package beego

import (
	"context"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func initSentinel(t *testing.T) {
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}

	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "GET:/ping",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "GET:/test",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "GET:/block",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "customize_block_fallback",
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

func TestSentinelMiddleware(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		//handler http.Handler
		handlerFunc web.HandleFunc
		body        io.Reader
	}
	type want struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default get",
			args: args{
				opts:    []Option{},
				method:  http.MethodGet,
				path:    "/ping",
				reqPath: "/ping",
				handlerFunc: func(ctx *beegoCtx.Context) {
					_ = ctx.Resp([]byte("pong"))
				},
				body: nil,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "customize resource extract",
			args: args{
				opts: []Option{
					WithResourceExtractor(func(r *http.Request) string {
						return "customize_block_fallback"
					}),
				},
				method:  http.MethodPost,
				path:    "/ping",
				reqPath: "/ping",
				handlerFunc: func(ctx *beegoCtx.Context) {
					_ = ctx.Resp([]byte("pong"))
				},
				body: nil,
			},
			want: want{
				code: http.StatusTooManyRequests,
			},
		},
		{
			name: "customize block fallback",
			args: args{
				opts: []Option{
					WithBlockFallback(func(r *http.Request) (int, string) {
						return http.StatusInternalServerError, "customize block fallback"
					}),
				},
				method:  http.MethodGet,
				path:    "/block",
				reqPath: "/block",
				handlerFunc: func(ctx *beegoCtx.Context) {
					_ = ctx.Resp([]byte("pong"))
				},
				body: nil,
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}

	initSentinel(t)
	defer func() {
		_ = flow.ClearRules()
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			middleware := SentinelMiddleware(tt.args.opts...)

			server := web.NewHttpSever()
			defer func() {
				_ = server.Server.Shutdown(context.Background())
			}()

			server.Get(tt.args.reqPath, tt.args.handlerFunc)
			server.Handlers.Init()
			server.Server.Handler = middleware(server.Handlers)

			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			w := httptest.NewRecorder()

			server.Server.Handler.ServeHTTP(w, r)

			assert.Equal(t, tt.want.code, w.Code)
		})
	}
}
