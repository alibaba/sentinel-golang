package stat

import "github.com/alibaba/sentinel-golang/core/base"

const (
	KeyMonitorBlockNode = "monitorBlockNode"
	KeyControlBlockNode = "controlBlockNode"
	KeyIsMonitorBlocked = "isMonitorBlocked"
	KeyMonitorBlockRule = "monitorBlockRule"
	KeyChildNodes       = "childNodes"
)

// SetFirstMonitorNodeAndRule set monitor block node. only record the first.
func SetFirstMonitorNodeAndRule(ctx *base.EntryContext, node *ResourceNode, rule base.SentinelRule) {
	if ctx == nil || node == nil {
		return
	}
	if IsMonitorBlocked(ctx) {
		return
	}

	PutOutputAttachment(ctx, KeyIsMonitorBlocked, true)
	PutOutputAttachment(ctx, KeyMonitorBlockNode, node)
	PutOutputAttachment(ctx, KeyMonitorBlockRule, rule)
}

func GetMonitorBlockInfo(ctx *base.EntryContext) (node *ResourceNode, rule base.SentinelRule, isMonitorBlocked bool) {
	if !IsMonitorBlocked(ctx) {
		return nil, nil, false
	}
	node, _ = GetOutputAttachment(ctx, KeyMonitorBlockNode).(*ResourceNode)
	rule, _ = GetOutputAttachment(ctx, KeyMonitorBlockRule).(base.SentinelRule)
	return node, rule, true
}

// IsMonitorBlocked reports whether current request was blocked in monitor mode.
func IsMonitorBlocked(ctx *base.EntryContext) bool {
	if ctx == nil {
		return false
	}
	isMonitorBlocked, ok := GetOutputAttachment(ctx, KeyIsMonitorBlocked).(bool)
	if !ok {
		return false
	}
	return isMonitorBlocked
}

func GetOutputAttachment(ctx *base.EntryContext, key interface{}) interface{} {
	if ctx.Data == nil {
		return nil
	}
	return ctx.Data[key]
}

func PutOutputAttachment(ctx *base.EntryContext, key interface{}, value interface{}) {
	if ctx.Data == nil {
		ctx.Data = make(map[interface{}]interface{})
	}
	ctx.Data[key] = value
}

// GetAllNodes gets all child node and parent node list.
func GetAllNodes(ctx *base.EntryContext) []*ResourceNode {
	childNodes, childOk := GetOutputAttachment(ctx, KeyChildNodes).([]*ResourceNode)
	var allNodes []*ResourceNode
	if childOk {
		allNodes = append(allNodes, childNodes...)
	}
	parentResNode, ok := ctx.StatNode.(*ResourceNode)
	if ok {
		allNodes = append(allNodes, parentResNode)
	}
	return allNodes
}

func SetBlockNode(ctx *base.EntryContext, node *ResourceNode) {
	PutOutputAttachment(ctx, KeyControlBlockNode, node)
}
