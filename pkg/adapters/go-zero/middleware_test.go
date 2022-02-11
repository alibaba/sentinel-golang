package go_zero

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
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
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "GET:/",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "/from/me",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "GET:/from/you",
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

type Request struct {
	Name string `path:"name,options=you|me"`
}

// go test -run ^TestSentinelMiddlewareDefault -v
func TestSentinelMiddlewareDefault(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler http.HandlerFunc
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
					path:    "/",
					reqPath: "/",
					handler: func(w http.ResponseWriter, r *http.Request) {
						resp := "index page"
						httpx.OkJson(w, &resp)
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
			var c rest.RestConf
			conf.MustLoad("./test.yml", &c)
			c.Port, _ = getAvailablePort(c.Port)
			server := rest.MustNewServer(c)
			// global middleware
			server.Use(SentinelMiddleware(tt.args.opts...))
			server.AddRoutes(
				[]rest.Route{
					{
						Method:  tt.args.method,
						Path:    tt.args.path,
						Handler: tt.args.handler,
					},
				},
			)
			go server.Start()
			defer server.Stop()
			time.Sleep(time.Duration(2) * time.Second)
			r, _ := http.Get(fmt.Sprintf("http://localhost:%d%s", c.Port, tt.args.reqPath))
			assert.Equal(t, tt.want.code, r.StatusCode)
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
		handler http.HandlerFunc
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
						WithResourceExtractor(func(r *http.Request) string {
							return r.URL.Path
						}),
					},
					method:  http.MethodGet,
					path:    "/from/:name",
					reqPath: "/from/me",
					handler: func(w http.ResponseWriter, r *http.Request) {
						var req Request
						if err := httpx.Parse(r, &req); err != nil {
							httpx.Error(w, err)
							return
						}
						resp := "from go-zero"
						httpx.OkJson(w, &resp)
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
			var c rest.RestConf
			conf.MustLoad("./test.yml", &c)
			c.Port, _ = getAvailablePort(c.Port)
			server := rest.MustNewServer(c)
			// global middleware
			server.Use(SentinelMiddleware(tt.args.opts...))
			server.AddRoutes(
				[]rest.Route{
					{
						Method:  tt.args.method,
						Path:    tt.args.path,
						Handler: tt.args.handler,
					},
				},
			)
			go server.Start()
			server.Stop()
			time.Sleep(time.Duration(2) * time.Second)
			r, _ := http.Get(fmt.Sprintf("http://localhost:%d%s", c.Port, tt.args.reqPath))
			assert.Equal(t, tt.want.code, r.StatusCode)
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
		handler http.HandlerFunc
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
						WithBlockFallback(func(r *http.Request) (int, string) {
							return http.StatusBadRequest, "Blocked with customized fallback"
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(w http.ResponseWriter, r *http.Request) {
						resp := "ping"
						httpx.OkJson(w, &resp)
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
			var c rest.RestConf
			conf.MustLoad("./test.yml", &c)
			c.Port, _ = getAvailablePort(c.Port)
			server := rest.MustNewServer(c)
			// global middleware
			server.Use(SentinelMiddleware(tt.args.opts...))
			server.AddRoutes(
				[]rest.Route{
					{
						Method:  tt.args.method,
						Path:    tt.args.path,
						Handler: tt.args.handler,
					},
				},
			)
			go server.Start()
			defer server.Stop()
			time.Sleep(time.Duration(2) * time.Second)
			r, _ := http.Get(fmt.Sprintf("http://localhost:%d%s", c.Port, tt.args.reqPath))
			assert.Equal(t, tt.want.code, r.StatusCode)
		})
	}
}

// go test -run ^TestSentinelMiddlewareRouting -v
func TestSentinelMiddlewareRouting(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler http.HandlerFunc
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
				name: "routing middleware",
				args: args{
					opts: []Option{
						WithResourceExtractor(func(r *http.Request) string {
							// global middleware won't block it in this way
							return ""
						}),
					},
					method:  http.MethodGet,
					path:    "/from/:name",
					reqPath: "/from/you",
					handler: func(w http.ResponseWriter, r *http.Request) {
						var req Request
						if err := httpx.Parse(r, &req); err != nil {
							httpx.Error(w, err)
							return
						}
						resp := "from go-zero"
						httpx.OkJson(w, &resp)
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
			var c rest.RestConf
			conf.MustLoad("./test.yml", &c)
			c.Port, _ = getAvailablePort(c.Port)
			server := rest.MustNewServer(c)
			// this `AddRoutes` is only for testing,
			// in practice, routing will be automatically carried by go-zero,
			// do not modify it in your application if you use goctl.
			server.AddRoutes(
				rest.WithMiddlewares(
					// routing middleware,
					[]rest.Middleware{NewSentinelRouteMiddleware().Handle},
					[]rest.Route{
						{
							Method:  tt.args.method,
							Path:    tt.args.path,
							Handler: tt.args.handler,
						},
					}...,
				),
			)
			go server.Start()
			defer server.Stop()
			time.Sleep(time.Duration(2) * time.Second)
			r, _ := http.Get(fmt.Sprintf("http://localhost:%d%s", c.Port, tt.args.reqPath))
			assert.Equal(t, tt.want.code, r.StatusCode)
		})
	}

}

func getAvailablePort(init int) (int, error) {
	for p := init; p < 65536; p++ {
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", fmt.Sprint(p)), time.Second)
		if conn != nil {
			conn.Close()
		} else {
			return p, nil
		}
	}
	return 0, fmt.Errorf("Cannot get an available port")
}
