package base

import (
	"github.com/sentinel-group/sentinel-golang/core/node"
)

type Entry struct {
	createTime   uint64
	originNode   node.Node
	currentNode  node.Node
	resourceWrap *ResourceWrapper
	err          error
}
