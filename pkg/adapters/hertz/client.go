package hertz

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// SentinelClientMiddleware returns new client.Middleware
// Default resource name is {method}:{path}, such as "GET:/api/users"
// Default block fallback is returning blockError
// Define your own behavior by setting serverOptions
func SentinelClientMiddleware(opts ...ClientOption) client.Middleware {
	options := newClientOptions(opts)
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
			resourceName := options.resourceExtract(ctx, req, resp)
			entry, blockErr := sentinel.Entry(
				resourceName,
				sentinel.WithResourceType(base.ResTypeWeb),
				sentinel.WithTrafficType(base.Outbound),
			)
			if blockErr != nil {
				return options.blockFallback(ctx, req, resp, blockErr)
			}

			defer entry.Exit()
			err = next(ctx, req, resp)
			if err != nil {
				sentinel.TraceError(entry, err)
				return err
			}
			return nil
		}
	}
}
