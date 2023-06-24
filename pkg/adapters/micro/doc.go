/*
This package provides Sentinel integration for go-micro.

For server side, users may append a Sentinel handler wrapper to go-micro service, like:

	import (
		sentinelPlugin "github.com/sentinel-go/pkg/adapters/micro"
	)

	// Append a Sentinel handler wrapper.
	micro.NewService(micro.WrapHandler(sentinelPlugin.NewHandlerWrapper()))

The plugin extracts service method as the resource name by default.
Users may provide customized resource name extractor when creating new
Sentinel handler wrapper (via options).

Fallback logic: the plugin will return the BlockError by default
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithXxxBlockFallback(handler) options.
*/
package micro
