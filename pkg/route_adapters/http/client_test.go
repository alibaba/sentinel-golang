package http

import (
	"context"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/route"
	"net/http"
	"testing"
)

func TestClientRoute(t *testing.T) {
	getTrafficTag := func(ctx context.Context) string {
		return ""
	}
	getPodTag := func(ctx context.Context) string {
		return "base"
	}
	setTrafficTag := func(ctx context.Context, trafficTag string) (context.Context, error) {
		return ctx, nil
	}

	err := route.NewCallbackFunc(nil, getTrafficTag, getPodTag, setTrafficTag)

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://gin-server-b.default.svc.cluster.local/greet", nil)
	if err != nil {
		fmt.Printf("[TestClientRoute] failed to create request: %v\n", err)
		return
	}

	fmt.Printf("[TestClientRoute] req before client route: %v\n", *req)
	req, err = ClientRoute(req)
	if err != nil {
		fmt.Printf("[TestClientRoute] failed to run client route: %v\n", err)
		t.Fail()
		return
	}
	fmt.Printf("[TestClientRoute] req after client route: %v\n", *req)
}
