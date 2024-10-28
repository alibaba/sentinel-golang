package kitex

import (
	"context"
	"fmt"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	ruleBasedResolver "github.com/kitex-contrib/resolver-rule-based"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/outlier"
)

var filterNodes []string
var halfNodes []string

// SentinelClientMiddleware returns new client.Middleware
// Default resource name is {service's name}:{method}
// Default block fallback is returning blockError
// Define your own behavior by setting serverOptions
func SentinelClientMiddleware(opts ...Option) func(endpoint.Endpoint) endpoint.Endpoint {
	options := newOptions(opts)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			if !options.EnableOutlier(ctx) {
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
			} else { // returns new client middleware specifically for outlier ejection.
				slotChain := sentinel.BuildDefaultSlotChain()
				slotChain.AddRuleCheckSlot(outlier.DefaultSlot)
				slotChain.AddStatSlot(outlier.DefaultMetricStatSlot)
				resourceName := ServiceNameExtract(ctx)
				entry, _ := sentinel.Entry(
					resourceName,
					sentinel.WithResourceType(base.ResTypeRPC),
					sentinel.WithTrafficType(base.Outbound),
					sentinel.WithSlotChain(slotChain),
				)
				defer entry.Exit()
				filterNodes = entry.Context().FilterNodes()
				halfNodes = entry.Context().HalfOpenNodes()
				err := next(ctx, req, resp)
				if callee := CalleeAddressExtract(ctx); callee != "" {
					sentinel.TraceCallee(entry, callee)
					if err != nil {
						sentinel.TraceError(entry, err)
					}
				}
				return err
			}
		}
	}
}

func OutlierClientResolver(resolver discovery.Resolver) discovery.Resolver {
	filterFunc := func(ctx context.Context, nodes []discovery.Instance) []discovery.Instance {
		var nodesPost []discovery.Instance
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
	// Construct the filterRule and build rule based resolver
	filterRule := &ruleBasedResolver.FilterRule{
		Name:  "outlier_filter_rule",
		Funcs: []ruleBasedResolver.FilterFunc{filterFunc},
	}
	return ruleBasedResolver.NewRuleBasedResolver(resolver, filterRule)
}

func getRemainingNodes(nodes []discovery.Instance, filters []string, flag bool) []discovery.Instance {
	nodesMap := make(map[string]struct{})
	for _, node := range filters {
		nodesMap[node] = struct{}{}
	}
	nodesPost := make([]discovery.Instance, 0)
	for _, ep := range nodes {
		if _, ok := nodesMap[ep.Address().String()]; ok == flag {
			nodesPost = append(nodesPost, ep)
		}
	}
	return nodesPost
}

// TODO remove this func
func printNodes(nodes []discovery.Instance) (res []string) {
	for _, v := range nodes {
		res = append(res, v.Address().String())
	}
	return
}
