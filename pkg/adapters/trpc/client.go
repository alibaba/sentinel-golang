package trpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"trpc.group/trpc-go/trpc-go/filter"
)

// SentinelClientFilter returns a new trpc client filter wrapped with Sentinel entry.
// Default resource name is {CalleeServiceName}:{ClientRPCName}.
func SentinelClientFilter(opts ...Option) filter.ClientFilter {
	options := newOptions(opts)
	return func(ctx context.Context, req, rsp interface{}, next filter.ClientHandleFunc) error {
		resourceName := options.resourceExtract(ctx, req)
		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)
		if blockErr != nil {
			return blockErr
		}
		defer entry.Exit()

		err := next(ctx, req, rsp)
		if err != nil {
			sentinel.TraceError(entry, err)
		}
		return err
	}
}
