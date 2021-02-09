package gear

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/stretchr/testify/assert"
	"github.com/teambition/gear"
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
		handler func(ctx *gear.Context) error
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
					handler: func(ctx *gear.Context) error {
						return ctx.End(http.StatusOK, []byte("ping"))
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
						WithResourceExtractor(func(ctx *gear.Context) string {
							return gear.GetRouterPatternFromCtx(ctx)
						}),
					},
					method:  http.MethodPost,
					path:    "/api/users/:id",
					reqPath: "/api/users/123",
					handler: func(ctx *gear.Context) error {
						return ctx.End(http.StatusOK, []byte("ping"))
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
						WithBlockFallback(func(ctx *gear.Context) error {
							return ctx.End(http.StatusBadRequest, []byte("block"))
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(ctx *gear.Context) error {
						return ctx.End(http.StatusOK, []byte("ping"))
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
			app := gear.New()
			router := gear.NewRouter()
			router.Use(SentinelMiddleware(tt.args.opts...))
			router.Handle(tt.args.method, tt.args.path, tt.args.handler)
			app.UseHandler(router)
			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, r)
			t.Log(tt.name)
			assert.Equal(t, tt.want.code, w.Code)
		})
	}
}
