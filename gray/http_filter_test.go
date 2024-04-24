package gray

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/baggage"
	"net/http"
	"testing"
)

func TestGrayOutboundFilterHttp(t *testing.T) {
	member, err := baggage.NewMember(baggageGrayTag, "gray")
	if err != nil {
		fmt.Printf("[GrayOutboundFilterHttp] new member err: %v, member: %v\n", err, member)
		t.Error(err)
		return
	}

	bag, err := baggage.New(member)
	if err != nil {
		fmt.Printf("[GrayOutboundFilterHttp] set member err: %v, member: %v\n", err, member)
		t.Error(err)
		return
	}
	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	//ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://grpc-server-c/greet", nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("version", "v1")

	GrayOutboundFilterHttp(req)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("[TestGratyOutboundFilterHttp] http resp: %v\n", resp)
}
