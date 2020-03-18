package rocketmq

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// SentinelConsumerInterceptor returns interceptor for consumer
func SentinelConsumerInterceptor(opts ...Option) primitive.Interceptor {
	options := evaluateOptions(opts)
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		consumerCtx, _ := primitive.GetConsumerCtx(ctx)
		resourceName := consumerCtx.ConsumerGroup + ":" + consumerCtx.MQ.Topic

		if options.consumerResourceExtract != nil {
			resourceName = options.consumerResourceExtract(consumerCtx)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeMQ),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				return options.blockFallback(ctx, req, reply, next, err)
			} else {
				return err
			}
		}

		defer entry.Exit()

		return next(ctx, req, reply)
	}
}
