/*
This package provides Sentinel integration for go-zero.

go-zero provides both zrpc and rest modules.

For go-zero/zrpc, user can call `AddUnaryInterceptors` and `AddStreamInterceptors`
on zrpc.RpcServer or `WithUnaryClientInterceptor` on zrpc.ClientOption to add grpc interceptors.

For go-zero/rest, there are two kinds of middlewares.
The first one is the global middleware, which can be applied to rest.Server via:

	import (
		sgz "github.com/sentinel-go/pkg/adapters/go-zero"
	)
	server := rest.MustNewServer(c.RestConf)
	server.Use(sgz.SentinelMiddleware())

The plugin extracts service FullMethod as the resource name by default.
Users may provide customized resource name extractor when creating new
Sentinel interceptors (via options).

Fallback logic: the plugin will return the BlockError by default
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithXxxBlockFallback(handler) options.

The second one is the routing middleware,
which is registered to specific routes via by go-zero automatically.
Therefore, it is recomended to create routing middleware based on the generated templates.
An example with Sentinel based on go-zero template is provided in `routing_middleware`.

The calling order is first global ones, then routing ones.
*/
package go_zero
