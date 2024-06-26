/*
This package provides Sentinel integration for gRPC.

For server/client side, users may create a gRPC server/client with Sentinel interceptor.
A sample of server side:

	import (
		sentinelPlugin "github.com/sentinel-group/sentinel-go-adapters/grpc"
		"google.golang.org/grpc"
	)

	// Create with Sentinel interceptor
	s := grpc.NewServer(grpc.UnaryInterceptor(sentinelPlugin.NewUnaryServerInterceptor()))

The plugin extracts service FullMethod as the resource name by default.
Users may provide customized resource name extractor when creating new
Sentinel interceptors (via options).

Fallback logic: the plugin will return the BlockError by default
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithXxxBlockFallback(handler) options.
*/
package grpc
