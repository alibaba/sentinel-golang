/*
This package provides Sentinel integration for hertz

For server side, users may create a Hertz Server with Sentinel middleware.
A simple of server side:

	import (
		sentinelPlugin "github.com/sentinel-go/pkg/adapters/hertz"
		"github.com/cloudwego/hertz/pkg/app/server"
	)

	h := server.New()
	h.Use(sentinelPlugin.SentinelServerMiddleware())

For client side, users may create a Hertz Client with Sentinel middleware.
A simple of client side:

	import (
		sentinelPlugin "github.com/sentinel-go/pkg/adapters/hertz"
		"github.com/cloudwego/hertz/pkg/app/client"
	)

	client, _ := client.NewClient()
	client.Use(sentinelPlugin.SentinelClientMiddleware())

The plugin extracts service FullMethod as the resource name by default.
Users may provide customized resource name extractor when creating new
Sentinel middlewares (via options).

Fallback logic: the plugin will stop by default if
current request is blocked by Sentinel rules. Users may also provide
customized fallback logic via WithClientBlockFallback(handler) options
for client side.

the plugin will return "429 Too Many Requests" status code
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithServerBlockFallback(handler)
options for server side.
*/
package hertz
