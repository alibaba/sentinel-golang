package dubbo

import (
	"bytes"
	"fmt"

	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/protocol"
)

const (
	ProviderFilterName = "sentinel-provider"
	ConsumerFilterName = "sentinel-consumer"

	DefaultProviderPrefix = "dubbo:provider:"
	DefaultConsumerPrefix = "dubbo:consumer:"

	MethodEntryKey    = "dubboMethodEntry"
	InterfaceEntryKey = "dubboInterfaceEntry"
)

// Currently, a ConcurrentHashMap mechanism is missing.
// All values are filled with default values first.

func getResourceName(invoker protocol.Invoker, invocation protocol.Invocation, prefix string) string {
	var (
		buf               bytes.Buffer
		interfaceResource string
	)
	buf.WriteString(prefix)
	if getInterfaceGroupAndVersionEnabled() {
		interfaceResource = getColonSeparatedKey(invoker.GetUrl())
	} else {
		interfaceResource = invoker.GetUrl().Service()
	}
	buf.WriteString(interfaceResource)
	buf.WriteString(":")
	buf.WriteString(invocation.MethodName())
	buf.WriteString("(")
	isFirst := true
	for _, v := range invocation.ParameterTypes() {
		if !isFirst {
			buf.WriteString(",")
		}
		buf.WriteString(v.Name())
		isFirst = false
	}
	buf.WriteString(")")
	return buf.String()
}

func getConsumerPrefix() string {
	return DefaultConsumerPrefix
}

func getProviderPrefix() string {
	return DefaultProviderPrefix
}

func getInterfaceGroupAndVersionEnabled() bool {
	return true
}

func getColonSeparatedKey(url common.URL) string {
	return fmt.Sprintf("%s:%s:%s",
		url.Service(),
		url.GetParam(constant.GROUP_KEY, ""),
		url.GetParam(constant.VERSION_KEY, ""))
}
