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

func init() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func GrayOutboundFilterHttp(req *http.Request) *http.Request {
	// TODO: 具体使用哪个ot sdk需要确认,以及需要从真实baggage中解析标签
	// 解析并获取流量标签,如果不存在且当前节点为灰度节点,将流量标签更新为节点标签
	trafficTag, newCtx := updateTrafficTagWithPodTag(req.Context())

	req = req.WithContext(newCtx)
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	if trafficTag != "" {
		req.Header.Set(baggageGrayTag, trafficTag)
	}
	fmt.Printf("[GratyOutboundFilterHttp] after tag update by pod tag: %+v\n", *req)

	// rds匹配
	header := make(map[string]string)
	for k, v := range req.Header {
		if len(v) != 0 {
			header[k] = v[0]
		}
	}
	newHost, newPort, newTrafficTag, update, err := rewriteByRds(req.Method, req.URL.Hostname(), req.URL.Port(), req.URL.Path, req.URL.Scheme, header)
	if err != nil {
		fmt.Printf("[GrayOutboundFilterHttp] rewrite by rds err: %v, req: %+v\n", err, req)
	}
	if update {
		req.URL.Host = fmt.Sprintf("%s:%s", newHost, newPort)
		if newTrafficTag != "" {
			req = req.WithContext(updateTrafficTag(req.Context(), newTrafficTag))
			otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
			req.Header.Set(baggageGrayTag, newTrafficTag)
		}
		return req
	}

	// cds匹配
	newHost, newPort, err = rewriteByCds(trafficTag, req.URL.Hostname(), req.URL.Port(), req.URL.Scheme)
	if err != nil {
		fmt.Printf("[GrayOutboundFilterHttp] rewrite by cds err: %v, req: %+v\n", err, req)
		return req
	}
	req.URL.Host = fmt.Sprintf("%s:%s", newHost, newPort)
	return req
}

func updateTrafficTag(ctx context.Context, trafficTag string) context.Context {
	member, err := baggage.NewMember(baggageGrayTag, trafficTag)
	if err != nil {
		return ctx
	}

	bag := baggage.FromContext(ctx)
	bag, err = bag.SetMember(member)
	if err != nil {
		return ctx
	}

	return baggage.ContextWithBaggage(ctx, bag)
}

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

func rewriteByRds(method, host, port, path, scheme string, header map[string]string) (string, string, string, bool, error) {
	if port == "" {
		if scheme == "https" {
			port = httpsDefaultPort
		} else {
			port = httpDefaultPort
		}
	}

	return getRewriteHostByRds(method, host, port, path, header)
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
	newHost, newPort, err := getRewriteHostByCds(host, port, trafficTag)
	if err != nil {
		fmt.Printf("[rewriteByCds] rewrite address err: %v\n", err)
		return "", "", err
	}

	return newHost, newPort, nil
}
