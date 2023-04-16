package iris

import (
	"io"
	"net/http"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/httptest"
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

func TestSentinelMiddleware(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler func(ctx iris.Context)
		body    io.Reader
	}
	type want struct {
		code int
	}
	var (
		tests = []struct {
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
					handler: func(ctx iris.Context) {
						ctx.StatusCode(http.StatusOK)
						ctx.WriteString("ping")
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
						WithResourceExtractor(func(ctx iris.Context) string {
							return ctx.Request().URL.String()
						}),
					},
					method:  http.MethodPost,
					path:    "/api/users/:id",
					reqPath: "/api/users/123",
					handler: func(ctx iris.Context) {
						ctx.StatusCode(http.StatusOK)
						ctx.WriteString("ping")
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
						WithBlockFallback(func(ctx iris.Context) {
							ctx.StatusCode(http.StatusBadRequest)
							ctx.WriteString("block")
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(ctx iris.Context) {
						ctx.StatusCode(http.StatusOK)
						ctx.WriteString("ping")
					},
					body: nil,
				},
				want: want{
					code: http.StatusBadRequest,
				},
			},
		}
	)
	initSentinel(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := iris.New()
			router.Use(SentinelMiddleware(tt.args.opts...))
			router.Handle(tt.args.method, tt.args.path, tt.args.handler)

			e := httptest.New(t, router)

			if tt.args.method == http.MethodGet {
				e.GET(tt.args.path).
					Expect().
					Status(tt.want.code)
			} else if tt.args.method == http.MethodPost {
				e.POST(tt.args.path).
					Expect().
					Status(tt.want.code)
			}
		})
	}
}
