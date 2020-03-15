package beego

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.M) {
	_ = sentinel.InitDefault()
	t.Run()
}

func TestSentinelController(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := beego.NewControllerRegister()
	handler.Add("/user/:id", NewSentinelController(&MainController{}))

	_, _ = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "GET:/user/:id",
			MetricType:      flow.QPS,
			Count:           1,
			ControlBehavior: flow.Reject,
		},
	})

	r, _ := http.NewRequest("GET", "/user/123", nil)
	handler.ServeHTTP(recorder, r)
	assert.Equal(t, 200, recorder.Code)

	r, _ = http.NewRequest("GET", "/user/124", nil)
	handler.ServeHTTP(recorder, r)
	assert.Equal(t, 429, recorder.Code)
}

func TestSentinelFilterFunc(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := beego.NewControllerRegister()
	handler.Get("/user/:id", NewSentinelFilterFunc(func(ctx *context.Context) {
		ctx.WriteString("/users/" + ctx.Input.Param("id"))
	}))

	_, _ = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "GET:/user/:id",
			MetricType:      flow.QPS,
			Count:           1,
			ControlBehavior: flow.Reject,
		},
	})

	r, _ := http.NewRequest("GET", "/user/123", nil)
	handler.ServeHTTP(recorder, r)
	assert.Equal(t, 200, recorder.Code)

	r, _ = http.NewRequest("GET", "/user/124", nil)
	handler.ServeHTTP(recorder, r)
	assert.Equal(t, 200, recorder.Code)
}
