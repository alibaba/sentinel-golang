/*
This package provides Sentinel middleware for Beego.

Users may register SentinelMiddleware to the Beego server, like.

	import (
		sentinelPlugin "github.com/alibaba/sentinel-golang/pkg/adapters/beego"
		"github.com/beego/beego/v2/server/web"
	)

	web.RunWithMiddleWares(":0", sentinelPlugin.SentinelMiddleware())

The plugin uses the HTTP method and raw request URL path as the resource name by default (e.g. GET:/api/users/123).
The path is Request.URL.Path (decoded, without the query string), rather than
the registered route pattern (e.g. /api/users/:id).
Users may provide customized resource name extractor when creating new
SentinelMiddleware (via options).

Fallback logic: the plugin will return "429 Too Many Requests" status code
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithBlockFallback(handler) options.
*/
package beego
