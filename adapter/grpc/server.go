package grpc

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

// SentinelUnaryServerIntercept implements gRPC unary server interceptor interface
func SentinelUnaryServerIntercept(opts ...Option) grpc.UnaryServerInterceptor {
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
		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if err != nil {
			if options.unaryServerBlockFallback != nil {
				return options.unaryServerBlockFallback(ctx, req, info, err)
			}
			return nil, err
		}
		defer entry.Exit()
		return handler(ctx, req)
	}
}

// SentinelStreamServerIntercept implements gRPC stream server interceptor interface
func SentinelStreamServerIntercept(opts ...Option) grpc.StreamServerInterceptor {
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
		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)
		if err != nil { // blocked
			if options.streamServerBlockFallback != nil {
				return options.streamServerBlockFallback(srv, ss, info, err)
			}
			return err
		}
		defer entry.Exit()
		return handler(srv, ss)
	}
}
