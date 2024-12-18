package goframe

import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
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
			Resource:               "GET:/",
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
		{
			Resource:               "GET:/ping",
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

// go test -run ^TestSentinelMiddlewareDefault -v
func TestSentinelMiddlewareDefault(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler func(r *ghttp.Request)
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
					opts: []Option{
						WithResourceExtractor(func(r *ghttp.Request) string {
							return r.Router.Uri
						}),
					},
					method:  http.MethodPost,
					path:    "/",
					reqPath: "/",
					handler: func(r *ghttp.Request) {
						r.Response.WriteStatusExit(http.StatusOK, "/")
					},
					body: nil,
				},
				want: want{
					code: http.StatusOK,
				},
			},
		}
	)

	initSentinel(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := g.Server()
			s.SetRouteOverWrite(true)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(SentinelMiddleware(tt.args.opts...))
				group.ALL(tt.args.path, tt.args.handler)
			})
			s.Start()

			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)

			assert.Equal(t, tt.want.code, w.Code)
		})
	}
}

// go test -run ^TestSentinelMiddlewareExtractor -v
func TestSentinelMiddlewareExtractor(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler func(r *ghttp.Request)
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
				name: "customize resource extract",
				args: args{
					opts: []Option{
						WithResourceExtractor(func(r *ghttp.Request) string {
							return r.Router.Uri
						}),
					},
					method:  http.MethodPost,
					path:    "/api/users/:id",
					reqPath: "/api/users/123",
					handler: func(r *ghttp.Request) {
						r.Response.WriteStatusExit(http.StatusOK, "/api/users/123")
					},
					body: nil,
				},
				want: want{
					code: http.StatusTooManyRequests,
				},
			},
		}
	)

	initSentinel(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := g.Server()
			s.SetRouteOverWrite(true)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(SentinelMiddleware(tt.args.opts...))
				group.ALL(tt.args.path, tt.args.handler)
			})
			s.Start()

			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			assert.Equal(t, tt.want.code, w.Code)
		})
	}
}

// go test -run ^TestSentinelMiddlewareFallback -v
func TestSentinelMiddlewareFallback(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler func(r *ghttp.Request)
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
				name: "customize block fallback",
				args: args{
					opts: []Option{
						WithBlockFallback(func(r *ghttp.Request) {
							r.Response.WriteStatus(http.StatusBadRequest, "/ping")
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(r *ghttp.Request) {
						r.Response.WriteStatus(http.StatusOK, "ping")
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
			s := g.Server()
			s.SetRouteOverWrite(true)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(SentinelMiddleware(tt.args.opts...))
				group.ALL(tt.args.path, tt.args.handler)
			})
			s.Start()

			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			assert.Equal(t, tt.want.code, w.Code)
		})
	}
}
