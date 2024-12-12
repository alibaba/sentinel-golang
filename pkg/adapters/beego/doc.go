/*
This package provides Sentinel middleware for Beego.

Users may register SentinelMiddleware to the Beego server, like.

	import (
		sentinelPlugin "github.com/sentinel-group/sentinel-go-adapters/beego"
		"github.com/beego/beego/v2/server/web"
	)

	web.RunWithMiddleWares(":0", middleware)

The plugin extracts "HttpMethod:FullPath" as the resource name by default (e.g. GET:/foo/:id).
Users may provide customized resource name extractor when creating new
SentinelMiddleware (via options).

Fallback logic: the plugin will return "429 Too Many Requests" status code
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithBlockFallback(handler) options.
*/
package beego
