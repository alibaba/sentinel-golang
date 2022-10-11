package fiber

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
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
			Resource:               "/api/users/123",
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
		handler func(ctx *fiber.Ctx) error
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
					handler: func(ctx *fiber.Ctx) error {
						return ctx.Send([]byte("ping"))
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
						WithResourceExtractor(func(ctx *fiber.Ctx) string {
							return string(ctx.Context().Path())
						}),
					},
					method:  http.MethodPost,
					path:    "/api/users/:id",
					reqPath: "/api/users/123",
					handler: func(ctx *fiber.Ctx) error {
						return ctx.Send([]byte("ping"))
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
						WithBlockFallback(func(ctx *fiber.Ctx) error {
							return ctx.Status(http.StatusBadRequest).Send([]byte("block"))
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(ctx *fiber.Ctx) error {
						return ctx.Status(http.StatusOK).Send([]byte("ping"))
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
			app := fiber.New()
			app.Use(SentinelMiddleware(tt.args.opts...))
			app.Add(tt.args.method, tt.args.path, tt.args.handler)
			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			resp, err := app.Test(r)
			//body, _ := io.ReadAll(resp.Body)
			//fmt.Println(string(body))
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
