package gray

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"os"
)

func updateTrafficTagWithPodTag(ctx context.Context) (string, context.Context) {
	bag := baggage.FromContext(ctx)
	trafficTag := bag.Member(baggageGrayTag).Value()
	if trafficTag == "" {
		podGrayTag := os.Getenv(envPodGrayTag)
		if podGrayTag != "" && podGrayTag != baseVersion {
			trafficTag = podGrayTag
			member, err := baggage.NewMember(baggageGrayTag, podGrayTag)
			if err != nil {
				return "", ctx
			}

			bag, err = bag.SetMember(member)
			if err != nil {
				return "", ctx
			}

			return trafficTag, baggage.ContextWithBaggage(ctx, bag)
		}
	}

	return trafficTag, ctx
}

func rewriteByCds(trafficTag, host, port, scheme string) (string, string, error) {
	if trafficTag == "" {
		trafficTag = baseVersion
	}
	if port == "" {
		if scheme == "https" {
			port = httpsDefaultPort
		} else {
			port = httpDefaultPort
		}
	}

	fmt.Printf("[rewriteByCds] host: %v, port: %v, traffic tag: %v\n", host, port, trafficTag)
	newHost, newPort, err := getRewriteHost(host, port, trafficTag)
	if err != nil {
		fmt.Printf("[rewriteByCds] rewrite address err: %v\n", err)
		return "", "", err
	}

	return newHost, newPort, nil
}

func GrayOutboundFilterHttp(req *http.Request) {
	// TODO: 具体使用哪个ot sdk需要确认,以及需要从真实baggage中解析标签
	// 解析并获取流量标签,如果不存在且当前节点为灰度节点,将流量标签更新为节点标签
	trafficTag, newCtx := updateTrafficTagWithPodTag(req.Context())
	req = req.WithContext(newCtx)
	otel.GetTextMapPropagator().Inject(newCtx, propagation.HeaderCarrier(req.Header))

	// TODO: rds匹配:1.是否有匹配规则,2.从规则中解析标签和cluster name,3.覆盖式更新流量标签,4.根据cluster name获取节点

	// cds匹配
	newHost, newPort, err := rewriteByCds(trafficTag, req.URL.Hostname(), req.URL.Port(), req.URL.Scheme)
	if err != nil {
		fmt.Printf("[GrayOutboundFilterHttp] rewrite by cds err: %v, req: %+v\n", err, req)
		return
	}
	req.URL.Host = fmt.Sprintf("%s:%s", newHost, newPort)
}
