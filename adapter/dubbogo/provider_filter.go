package dubbogo

import (
	"github.com/apache/dubbo-go/protocol"
)
import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

type providerFilter struct{}

func (d *providerFilter) Invoke(invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	methodResourceName := getResourceName(invoker, invocation, getProviderPrefix())
	interfaceResourceName := ""
	if getInterfaceGroupAndVersionEnabled() {
		interfaceResourceName = getColonSeparatedKey(invoker.GetUrl())
	} else {
		interfaceResourceName = invoker.GetUrl().Service()
	}
	_, b := sentinel.Entry(interfaceResourceName,
		sentinel.WithResourceType(base.ResTypeRPC),
		sentinel.WithTrafficType(base.Inbound))
	if b != nil { // blocked
		result := &protocol.RPCResult{}
		result.SetResult(nil)
		result.SetError(b)
		return result
	}
	_, b = sentinel.Entry(methodResourceName,
		sentinel.WithResourceType(base.ResTypeRPC),
		sentinel.WithTrafficType(base.Inbound),
		sentinel.WithArgs(invocation.Attachments()))
	if b != nil { // blocked
		result := &protocol.RPCResult{}
		result.SetResult(nil)
		result.SetError(b)
		return result
	}

	return invoker.Invoke(invocation)
}

func (d *providerFilter) OnResponse(result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	return result
}
