package dubbogo

import (
	"github.com/apache/dubbo-go/filter"
)


func GetConsumerFilter() filter.Filter{
return &dubboConsumerFilter{}
}
