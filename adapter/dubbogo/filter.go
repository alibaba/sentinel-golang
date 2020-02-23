package dubbogo

import (
	"github.com/apache/dubbo-go/common/extension"
	"github.com/apache/dubbo-go/filter"
)

func init() {
	extension.SetFilter(ProviderFilterName, GetProviderFilter)
	extension.SetFilter(ConsumerFilterName, GetConsumerFilter)
}

func GetConsumerFilter() filter.Filter {
	return &consumerFilter{}
}

func GetProviderFilter() filter.Filter {
	return &providerFilter{}
}
