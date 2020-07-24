package dubbo

import (
	"context"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/apache/dubbo-go/protocol"
)

type consumerFilter struct{}

func (d *consumerFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	methodResourceName := getResourceName(invoker, invocation, getConsumerPrefix())
	interfaceResourceName := ""
	if getInterfaceGroupAndVersionEnabled() {
		interfaceResourceName = getColonSeparatedKey(invoker.GetUrl())
	} else {
		interfaceResourceName = invoker.GetUrl().Service()
	}
	var (
		interfaceEntry *base.SentinelEntry
		methodEntry    *base.SentinelEntry
		b              *base.BlockError
	)

	interfaceEntry, b = sentinel.Entry(interfaceResourceName, sentinel.WithResourceType(base.ResTypeRPC), sentinel.WithTrafficType(base.Outbound))
	if b != nil {
		// interface blocked
		return consumerDubboFallback(ctx, invoker, invocation, b)
	}
	ctx = context.WithValue(ctx, InterfaceEntryKey, interfaceEntry)

	methodEntry, b = sentinel.Entry(methodResourceName, sentinel.WithResourceType(base.ResTypeRPC),
		sentinel.WithTrafficType(base.Outbound), sentinel.WithArgs(invocation.Arguments()...))
	if b != nil {
		// method blocked
		return consumerDubboFallback(ctx, invoker, invocation, b)
	}
	ctx = context.WithValue(ctx, MethodEntryKey, methodEntry)

	return invoker.Invoke(ctx, invocation)
}

func (d *consumerFilter) OnResponse(ctx context.Context, result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	if methodEntry := ctx.Value(MethodEntryKey); methodEntry != nil {
		e := methodEntry.(*base.SentinelEntry)
		sentinel.TraceError(e, result.Error())
		e.Exit()
	}
	if interfaceEntry := ctx.Value(InterfaceEntryKey); interfaceEntry != nil {
		e := interfaceEntry.(*base.SentinelEntry)
		sentinel.TraceError(e, result.Error())
		e.Exit()
	}
	return result
}
