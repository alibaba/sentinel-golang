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

const filterNodesKey = "filterNodes"
const halfNodesKey = "halfNodes"

func OutlierClientFilter(ctx context.Context, nodes []selector.Node) []selector.Node {
	var filterNodes, halfNodes []string
	if v, ok := metadata.FromClientContext(ctx); ok {
		filterNodes = v.Values(filterNodesKey)
		halfNodes = v.Values(halfNodesKey)
	}
	var nodesPost []selector.Node
	if len(halfNodes) != 0 {
		fmt.Println("Half Filter Pre: ", printNodes(nodes))
		nodesPost = getRemainingNodes(nodes, halfNodes, true)
		fmt.Println("Half Filter Post: ", printNodes(nodesPost))
	} else {
		fmt.Println("Filter Pre: ", printNodes(nodes))
		nodesPost = getRemainingNodes(nodes, filterNodes, false)
		fmt.Println("Filter Post: ", printNodes(nodesPost))
	}
	return nodesPost
}

func OutlierClientMiddleware(src middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		resourceName := ServiceNameExtract(ctx)
		if resourceName == "" {
			return nil, fmt.Errorf("resource name is empty")
		}
		entry, _ := sentinelApi.Entry(
			resourceName,
			sentinelApi.WithResourceType(base.ResTypeRPC),
			sentinelApi.WithTrafficType(base.Outbound),
		)
		defer entry.Exit()

		if v, ok := metadata.FromClientContext(ctx); ok {
			filterNodes := entry.Context().FilterNodes()
			for _, node := range filterNodes {
				v.Add(filterNodesKey, node)
			}
			halfNodes := entry.Context().HalfOpenNodes()
			for _, node := range halfNodes {
				v.Add(halfNodesKey, node)
			}
		}

		res, err := src(ctx, req)
		if p, ok := selector.FromPeerContext(ctx); ok && p.Node != nil {
			sentinelApi.TraceCallee(entry, p.Node.Address())
			if err != nil {
				sentinelApi.TraceError(entry, err)
			}
		}
		return res, err
	}
}

func getRemainingNodes(nodes []selector.Node, filters []string, flag bool) []selector.Node {
	nodesMap := make(map[string]struct{})
	for _, node := range filters {
		nodesMap[node] = struct{}{}
	}
	nodesPost := make([]selector.Node, 0)
	for _, ep := range nodes {
		if _, ok := nodesMap[ep.Address()]; ok == flag {
			nodesPost = append(nodesPost, ep)
		}
	}
	return nodesPost
}

func printNodes(nodes []selector.Node) (res []string) {
	for _, v := range nodes {
		res = append(res, v.Address())
	}
	return
}
