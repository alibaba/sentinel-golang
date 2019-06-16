package public

import (
	"github.com/sentinel-group/sentinel-golang/core/node"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type Entry struct {
	createTime   uint64
	originNode   node.Node
	currentNode  node.Node
	resourceWrap *base.ResourceWrapper
	err          error
}
