package kitex

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	ruleBasedResolver "github.com/kitex-contrib/resolver-rule-based"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

// SentinelClientMiddleware returns new client.Middleware
// Default resource name is {service's name}:{method}
// Default block fallback is returning blockError
// Define your own behavior by setting serverOptions
func SentinelClientMiddleware(opts ...Option) func(endpoint.Endpoint) endpoint.Endpoint {
	options := newOptions(opts)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			resourceName := options.ResourceExtract(ctx, req, resp)
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeRPC),
				sentinel.WithTrafficType(base.Outbound),
			)
			if blockErr != nil {
				return options.BlockFallback(ctx, req, resp, blockErr)
			}
			defer entry.Exit()
			err := next(ctx, req, resp)
			if err != nil {
				sentinel.TraceError(entry, err)
			}
			return err
		}
	}
}

// OutlierClientMiddleware returns new client.Middleware specifically for outlier removal.
func OutlierClientMiddleware(opts ...Option) func(endpoint.Endpoint) endpoint.Endpoint {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			resourceName := ServiceNameExtract(ctx, req, resp)
			entry, _ := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeRPC),
				sentinel.WithTrafficType(base.Outbound),
			)
			defer entry.Exit()
			ctx = context.WithValue(ctx, "filterNodes", entry.Context().FilterNodes())
			err0 := next(ctx, req, resp)
			if callee := DefaultAddressExtract(ctx); callee != "" {
				sentinel.TraceCallee(entry, callee)
				if err0 != nil {
					sentinel.TraceError(entry, err0)
				}
			}
			return err0
		}
	}
}

func OutlierClientResolver(resolver discovery.Resolver) discovery.Resolver {
	filterFunc := func(ctx context.Context, nodes []discovery.Instance) []discovery.Instance {
		nodesMap := make(map[string]struct{})
		filterNodes := ctx.Value("filterNodes").([]string) // TODO 传不过来
		for _, node := range filterNodes {
			nodesMap[node] = struct{}{}
		}
		fmt.Println("Filter Pre: ", printNodes(nodes))
		nodesCopy := make([]discovery.Instance, 0)
		for _, ep := range nodes {
			if _, ok := nodesMap[ep.Address().String()]; !ok {
				nodesCopy = append(nodesCopy, ep)
			}
		}
		fmt.Println("Filter Post: ", printNodes(nodesCopy))
		return nodesCopy
	}
	// Construct the filterRule and build rule based resolver
	filterRule := &ruleBasedResolver.FilterRule{Name: "rule-name", Funcs: []ruleBasedResolver.FilterFunc{filterFunc}}
	return ruleBasedResolver.NewRuleBasedResolver(resolver, filterRule)
}

// TODO remove this func
func printNodes(nodes []discovery.Instance) (res []string) {
	for _, v := range nodes {
		res = append(res, v.Address().String())
	}
	return
}
