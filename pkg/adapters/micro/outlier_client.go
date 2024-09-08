package micro

import (
	"context"
	"fmt"
	"slices"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"

	sentinelApi "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

type outlierClientWrapper struct {
	client.Client
}

// NewOutlierClientWrapper returns a sentinel outlier client Wrapper.
func NewOutlierClientWrapper(opts ...Option) client.Wrapper {
	return func(c client.Client) client.Client {
		return &outlierClientWrapper{c}
	}
}

func (c *outlierClientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	entry, _ := sentinelApi.Entry(
		req.Service(),
		sentinelApi.WithResourceType(base.ResTypeRPC),
		sentinelApi.WithTrafficType(base.Outbound),
	)
	defer entry.Exit()
	opts = append(opts, WithSelectOption(entry))
	opts = append(opts, WithCallWrapper(entry))
	return c.Client.Call(ctx, req, rsp, opts...)
}

func (c *outlierClientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	entry, _ := sentinelApi.Entry(
		req.Service(),
		sentinelApi.WithResourceType(base.ResTypeRPC),
		sentinelApi.WithTrafficType(base.Outbound),
	)
	defer entry.Exit()
	opts = append(opts, WithSelectOption(entry))
	opts = append(opts, WithCallWrapper(entry))
	stream, err := c.Client.Stream(ctx, req, opts...)
	if err != nil {
		sentinelApi.TraceError(entry, err)
	}
	return stream, err
}

func WithSelectOption(entry *base.SentinelEntry) client.CallOption {
	return client.WithSelectOption(selector.WithFilter(
		func(old []*registry.Service) (new []*registry.Service) {
			filterNodes := entry.Context().FilterNodes()
			halfNodes := entry.Context().HalfOpenNodes()
			if len(halfNodes) != 0 {
				fmt.Println("Half Filter Pre: ", printNodes(old[0].Nodes))
				new = getRemainingNodes(old, halfNodes)
				fmt.Println("Half Filter Post: ", printNodes(new[0].Nodes))
			} else {
				fmt.Println("Filter Pre: ", printNodes(old[0].Nodes))
				new = getRemainingNodes(old, filterNodes)
				fmt.Println("Filter Post: ", printNodes(new[0].Nodes))
			}
			return new
		},
	))
}

func WithCallWrapper(entry *base.SentinelEntry) client.CallOption {
	return client.WithCallWrapper(func(f1 client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request,
			rsp interface{}, opts client.CallOptions) error {
			err := f1(ctx, node, req, rsp, opts)
			sentinelApi.TraceCallee(entry, node.Address)
			if err != nil {
				sentinelApi.TraceError(entry, err)
			}
			return err
		}
	})
}

func getRemainingNodes(old []*registry.Service, filters []string) []*registry.Service {
	nodesMap := make(map[string]struct{})
	for _, node := range filters {
		nodesMap[node] = struct{}{}
	}
	for _, service := range old {
		nodesCopy := slices.Clone(service.Nodes)
		service.Nodes = make([]*registry.Node, 0)
		for _, ep := range nodesCopy {
			if _, ok := nodesMap[ep.Address]; !ok {
				service.Nodes = append(service.Nodes, ep)
			}
		}
	}
	return old
}

func printNodes(nodes []*registry.Node) (res []string) {
	for _, v := range nodes {
		res = append(res, v.Address)
	}
	return
}
