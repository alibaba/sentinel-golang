package dubbogo

import (
	"context"
	"github.com/apache/dubbo-go/protocol"
	sentinel "github.com/sentinel-group/sentinel-golang/api"
	"github.com/sentinel-group/sentinel-golang/core/base"
)

type dubboConsumerFilter struct{}

func (d *dubboConsumerFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	methodResourceName := getResourceName(invoker, invocation)
	interfaceResourceName := ""
	if getInterfaceGroupAndVersionEnabled() {
		interfaceResourceName = getColonSeparatedKey(invoker.GetUrl())
	} else {
		interfaceResourceName = invoker.GetUrl().Service()
	}
	if !isAsync(invocation) {
		_, b := sentinel.Entry(interfaceResourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound))
		if b != nil { // blocked
			result := &protocol.RPCResult{}
			result.SetResult(nil)
			result.SetError(b)
			return result
		}
		_, b = sentinel.Entry(methodResourceName,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
			sentinel.WithArgs(invocation.Attachments()))
		if b != nil { // blocked
			result := &protocol.RPCResult{}
			result.SetResult(nil)
			result.SetError(b)
			return result
		}
	} else {
		// todo : Need to implement asynchronous current limiting
	}
	return invoker.Invoke(invocation)
}

func (d *dubboConsumerFilter) OnResponse(_ context.Context, result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	return result
}
