package micro

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/micro/go-micro/v2/server"
)

// NewHandlerWrapper returns a Handler Wrapper with Alibaba Sentinel breaker
func NewHandlerWrapper(sentinelOpts ...Option) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			opts := evaluateOptions(sentinelOpts)
			resourceName := req.Method()
			if opts.serverResourceExtract != nil {
				resourceName = opts.serverResourceExtract(ctx, req)
			}
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeRPC),
				sentinel.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				if opts.serverBlockFallback != nil {
					return opts.serverBlockFallback(ctx, req, blockErr)
				}
				return blockErr
			}
			defer entry.Exit()
			err := h(ctx, req, rsp)
			if err != nil {
				sentinel.TraceError(entry, err)
			}
			return err
		}
	}
}

func NewStreamWrapper(sentinelOpts ...Option) server.StreamWrapper {
	return func(stream server.Stream) server.Stream {
		opts := evaluateOptions(sentinelOpts)
		resourceName := stream.Request().Method()
		if opts.serverResourceExtract != nil {
			resourceName = opts.streamServerResourceExtract(stream)
		}
		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if blockErr != nil {
			if opts.serverBlockFallback != nil {
				return opts.streamServerBlockFallback(stream, blockErr)
			}

			stream.Send(blockErr)
			return stream
		}

		entry.Exit()
		return stream
	}
}
