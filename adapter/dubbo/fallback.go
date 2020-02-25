package dubbo

import (
	"context"

	"github.com/apache/dubbo-go/protocol"
)
import (
	"github.com/alibaba/sentinel-golang/core/base"
)

var (
	consumerDubboFallback = getDefaultDubboFallback()
	providerDubboFallback = getDefaultDubboFallback()
)

type DubboFallback func(context.Context, protocol.Invoker, protocol.Invocation, *base.BlockError) protocol.Result

func SetConsumerDubboFallback(f DubboFallback) {
	consumerDubboFallback = f
}
func SetProviderDubboFallback(f DubboFallback) {
	providerDubboFallback = f
}
func getDefaultDubboFallback() DubboFallback {
	return func(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation, blockError *base.BlockError) protocol.Result {
		result := &protocol.RPCResult{}
		result.SetResult(nil)
		result.SetError(blockError)
		return result
	}
}
