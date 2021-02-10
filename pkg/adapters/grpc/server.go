package grpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

// NewUnaryServerInterceptor creates the unary server interceptor wrapped with Sentinel entry.
func NewUnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	options := evaluateOptions(opts)
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// method as resource name by default
		resourceName := info.FullMethod
		if options.unaryServerResourceExtract != nil {
			resourceName = options.unaryServerResourceExtract(ctx, req, info)
		}
		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if blockErr != nil {
			if options.unaryServerBlockFallback != nil {
				return options.unaryServerBlockFallback(ctx, req, info, blockErr)
			}
			return nil, blockErr
		}
		defer entry.Exit()

		res, err := handler(ctx, req)
		if err != nil {
			sentinel.TraceError(entry, err)
		}
		return res, err
	}
}

// NewStreamServerInterceptor creates the unary stream interceptor wrapped with Sentinel entry.
func NewStreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	options := evaluateOptions(opts)
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// method as resource name by default
		resourceName := info.FullMethod
		if options.streamServerResourceExtract != nil {
			resourceName = options.streamServerResourceExtract(srv, ss, info)
		}
		entry, blockErr := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if blockErr != nil { // blocked
			if options.streamServerBlockFallback != nil {
				return options.streamServerBlockFallback(srv, ss, info, blockErr)
			}
			return blockErr
		}
		defer entry.Exit()

		err := handler(srv, ss)
		if err != nil {
			sentinel.TraceError(entry, err)
		}
		return err
	}
}
