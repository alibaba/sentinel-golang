package micro

import (
	"context"
	"fmt"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func WithSelectOption(entry *base.SentinelEntry) client.CallOption {
	return client.WithSelectOption(selector.WithFilter(
		func(old []*registry.Service) (new []*registry.Service) {
			filterNodes := entry.Context().FilterNodes()
			halfNodes := entry.Context().HalfOpenNodes()
			if len(halfNodes) != 0 {
				fmt.Println("Half Filter Pre: ", printNodes(old[0].Nodes))
				new = getRemainingNodes(old, halfNodes, true)
				fmt.Println("Half Filter Post: ", printNodes(new[0].Nodes))
			} else {
				fmt.Println("Filter Pre: ", printNodes(old[0].Nodes))
				new = getRemainingNodes(old, filterNodes, false)
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
			sentinel.TraceCallee(entry, node.Address)
			if err != nil {
				sentinel.TraceError(entry, err)
			}
			return err
		}
	})
}

func getRemainingNodes(old []*registry.Service, filters []string, flag bool) []*registry.Service {
	nodesMap := make(map[string]struct{})
	for _, node := range filters {
		nodesMap[node] = struct{}{}
	}
	for _, service := range old {
		nodesCopy := make([]*registry.Node, 0)
		for _, ep := range service.Nodes {
			nodesCopy = append(nodesCopy, ep)
		}
		service.Nodes = make([]*registry.Node, 0)
		for _, ep := range nodesCopy {
			if _, ok := nodesMap[ep.Address]; ok == flag {
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
