/*
This package provides Sentinel middleware for Gear.

Users may register SentinelMiddleware to the Gear router, like.

	import (
		sentinelPlugear "github.com/sentinel-group/sentinel-go-adapters/gear"
		"github.com/teambition/gear"
	)

	r := gear.NewRouter()

	r.Use(sentinelPlugear.SentinelMiddleware())

The plugear extracts "HttpMethod:Router" as the resource name by default (e.g. GET:/foo/:id).
Users may provide customized resource name extractor when creating new
SentinelMiddleware (via options).

Fallback logic: the plugear will return "429 Too Many Requests" status code
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithBlockFallback(handler) options.
*/
package gear
