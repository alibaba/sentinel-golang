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

func TestSentinelFilter(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := beego.NewControllerRegister()
	beforeExec, finishRouter := SentinelFilters()
	assert.Nil(t, handler.InsertFilter("/ping/:id", beego.BeforeExec, beforeExec, false))
	assert.Nil(t, handler.InsertFilter("/ping/:id", beego.FinishRouter, finishRouter, false))
	handler.Any("/ping/:id", func(ctx *context.Context) {
		ctx.Output.SetStatus(200)
	})

	_, _ = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "GET:/ping/:id",
			MetricType:      flow.QPS,
			Count:           0,
			ControlBehavior: flow.Reject,
		},
	})
	
	r, _ := http.NewRequest("GET", "/ping/123", nil)
	handler.ServeHTTP(recorder, r)
	assert.Equal(t, 200, recorder.Code)

	// r, _ = http.NewRequest("GET", "/ping/1234", nil)
	// handler.ServeHTTP(recorder, r)
	// assert.Equal(t, 200, recorder.Code)
}