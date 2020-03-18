package grpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

// SentinelUnaryClientIntercept returns new grpc.UnaryClientInterceptor instance
func SentinelUnaryClientIntercept(opts ...Option) grpc.UnaryClientInterceptor {
	options := evaluateOptions(opts)
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// method as resource name by default
		resourceName := method
		if options.unaryClientResourceExtract != nil {
			resourceName = options.unaryClientResourceExtract(ctx, method, req, cc)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)
		if err != nil {
			if options.unaryClientBlockFallback != nil {
				return options.unaryClientBlockFallback(ctx, method, req, cc, err)
			}
			return err
		}
		defer entry.Exit()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// SentinelStreamClientIntercept returns new grpc.StreamClientInterceptor instance
func SentinelStreamClientIntercept(opts ...Option) grpc.StreamClientInterceptor {
	options := evaluateOptions(opts)
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// method as resource name by default
		resourceName := method
		if options.streamClientResourceExtract != nil {
			resourceName = options.streamClientResourceExtract(ctx, desc, cc, method)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)
		if err != nil { // blocked
			if options.streamClientBlockFallback != nil {
				return options.streamClientBlockFallback(ctx, desc, cc, method, err)
			}
			return nil, err
		}

		defer entry.Exit()

		return streamer(ctx, desc, cc, method, opts...)
	}
}
