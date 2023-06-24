/*
*
This package provides Sentinel integration for Kitex.

For server side, users may append a Sentinel middleware to Kitex service, like:

	import (
		sentinelPlugin "github.com/alibaba/sentinel-golang/pkg/adapters/kitex"
	)
	srv := hello.NewServer(new(HelloImpl),server.WithMiddleware(SentinelServerMiddleware()))

The plugin extracts service name and service method as the resource name by default.
Users may provide customized resource name extractor when creating new
Sentinel middleware (via WithResourceExtract options).

Fallback logic: the plugin will return the BlockError by default
if current request is blocked by Sentinel rules. Users may also
provide customized fallback logic via WithBlockFallback(handler) options.
*/
package kitex
