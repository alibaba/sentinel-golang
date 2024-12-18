// Package goframe provides Sentinel middleware for GoFrame.
//
// Users may register SentinelMiddleware to the GoFrame server, like:
//
//	import (
//		sentinelPlugin "github.com/your-repo/goframe-sentinel-adapter"
//		"github.com/gogf/gf/v2/frame/g"
//		"github.com/gogf/gf/v2/net/ghttp"
//	)
//
//	s := g.Server()
//	s.Use(ghttp.MiddlewareHandlerFunc(sentinelPlugin.SentinelMiddleware()))
//
// The plugin extracts "HttpMethod:FullPath" as the resource name by default (e.g. GET:/foo/:id).
// Users may provide a customized resource name extractor when creating new SentinelMiddleware (via options).
//
// Fallback logic: the plugin will return "429 Too Many Requests" status code if the current request is blocked by Sentinel rules.
// Users may also provide customized fallback logic via WithBlockFallback(handler) options.
package goframe
