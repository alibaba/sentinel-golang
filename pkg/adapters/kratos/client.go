package kratos

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/selector"

	sentinelApi "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func OutlierClientFilter(ctx context.Context, nodes []selector.Node) []selector.Node {
	var filterNodes, halfNodes []string
	if v, ok := metadata.FromClientContext(ctx); ok {
		filterNodes = v.Values("filterNodes")
		halfNodes = v.Values("halfNodes")
	}

	if len(halfNodes) != 0 {
		nodesMap := make(map[string]struct{})
		for _, node := range halfNodes {
			nodesMap[node] = struct{}{}
		}
		fmt.Println("Half Filter Pre: ", printNodes(nodes))
		nodesCopy := make([]selector.Node, 0)
		for _, ep := range nodes {
			if _, ok := nodesMap[ep.Address()]; ok {
				nodesCopy = append(nodesCopy, ep)
			}
		}
		fmt.Println("Half Filter Post: ", printNodes(nodesCopy))
		return nodesCopy
	} else {
		nodesMap := make(map[string]struct{})
		for _, node := range filterNodes {
			nodesMap[node] = struct{}{}
		}
		fmt.Println("Filter Pre: ", printNodes(nodes))
		nodesCopy := make([]selector.Node, 0)
		for _, ep := range nodes {
			if _, ok := nodesMap[ep.Address()]; !ok {
				nodesCopy = append(nodesCopy, ep)
			}
		}
		fmt.Println("Filter Post: ", printNodes(nodesCopy))
		return nodesCopy
	}
}

func OutlierClientMiddleware(src middleware.Handler) middleware.Handler {
	resourceName := "my_rpc_service"
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		entry, _ := sentinelApi.Entry(
			resourceName,
			sentinelApi.WithResourceType(base.ResTypeRPC),
			sentinelApi.WithTrafficType(base.Outbound),
		)
		defer entry.Exit()

		if v, ok := metadata.FromClientContext(ctx); ok {
			filterNodes := entry.Context().FilterNodes()
			for _, node := range filterNodes {
				v.Add("filterNodes", node)
			}
			halfNodes := entry.Context().HalfOpenNodes()
			for _, node := range halfNodes {
				v.Add("halfNodes", node)
			}
		}

		res, err := src(ctx, req)

		if p, ok := selector.FromPeerContext(ctx); ok {
			sentinelApi.TraceCallee(entry, p.Node.Address())
			if err != nil {
				sentinelApi.TraceError(entry, err)
			}
		}
		return res, err
	}
}

func printNodes(nodes []selector.Node) (res []string) {
	for _, v := range nodes {
		res = append(res, v.Address())
	}
	return
}
