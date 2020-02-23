package dubbogo

import (
	"github.com/apache/dubbo-go/protocol"
)
import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

type consumerFilter struct{}

func (d *consumerFilter) Invoke(invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	methodResourceName := getResourceName(invoker, invocation, getConsumerPrefix())
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
		//  unlimited flow for the time being
	}
	return invoker.Invoke(invocation)
}

func (d *consumerFilter) OnResponse(result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	return result
}
