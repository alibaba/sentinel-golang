package trpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"trpc.group/trpc-go/trpc-go/filter"
)

// SentinelServerFilter returns a new trpc server filter wrapped with Sentinel entry.
// Default resource name is {CalleeServiceName}:{ServerRPCName}.
func SentinelServerFilter(opts ...Option) filter.ServerFilter {
	options := newOptions(opts)
	return func(ctx context.Context, req interface{}, next filter.ServerHandleFunc) (rsp interface{}, err error) {
		resourceName := options.resourceExtract(ctx, req)
		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if blockErr != nil {
			return options.blockFallback(ctx, req, blockErr)
		}
		defer entry.Exit()

		rsp, err = next(ctx, req)
		if err != nil {
			sentinel.TraceError(entry, err)
		}
		return rsp, err
	}
}
