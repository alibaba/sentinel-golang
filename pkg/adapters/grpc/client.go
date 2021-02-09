package grpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

// NewUnaryClientInterceptor creates the unary client interceptor wrapped with Sentinel entry.
func NewUnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
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

		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)
		if blockErr != nil {
			if options.unaryClientBlockFallback != nil {
				return options.unaryClientBlockFallback(ctx, method, req, cc, blockErr)
			}
			return blockErr
		}
		defer entry.Exit()

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			sentinel.TraceError(entry, err)
		}
		return err
	}
}

// NewStreamClientInterceptor creates the stream client interceptor wrapped with Sentinel entry.
func NewStreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
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

		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)
		if blockErr != nil { // blocked
			if options.streamClientBlockFallback != nil {
				return options.streamClientBlockFallback(ctx, desc, cc, method, blockErr)
			}
			return nil, blockErr
		}
		defer entry.Exit()

		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			sentinel.TraceError(entry, err)
		}

		return cs, err
	}
}
