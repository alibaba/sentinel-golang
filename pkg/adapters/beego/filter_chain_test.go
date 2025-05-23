package beego

import (
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/beego/beego/v2/server/web"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSentinelFilterChain(t *testing.T) {
	type args struct {
		opts        []Option
		method      string
		path        string
		reqPath     string
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
				path:    "/test",
				reqPath: "/test",
				handlerFunc: func(ctx *beegoCtx.Context) {
					_ = ctx.Resp([]byte("hello"))
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
				path:    "/api/users/:id",
				reqPath: "/api/users/123",
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
			cr := web.NewControllerRegister()
			cr.Get(tt.args.reqPath, tt.args.handlerFunc)

			cr.InsertFilterChain("/*", SentinelFilterChain(tt.args.opts...))
			cr.Init()

			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, tt.args.body)
			w := httptest.NewRecorder()

			cr.ServeHTTP(w, r)
		})
	}
}
