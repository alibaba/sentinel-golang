package fallback

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

var blockType2FuncTypeMap = map[base.BlockType]FunctionType{
	base.BlockTypeFlow:             FlowType,
	base.BlockTypeIsolation:        Isolation,
	base.BlockTypeHotSpotParamFlow: HotspotRpc,
}

func getFuncTypeByBlockType(blockType base.BlockType) (FunctionType, bool) {
	funcType, ok := blockType2FuncTypeMap[blockType]
	return funcType, ok
}

func GetWebFallbackBehavior(resource string, blockType base.BlockType) (*WebBlockFallbackBehavior, bool) {
	functionType, ok := getFuncTypeByBlockType(blockType)
	if !ok {
		return nil, false
	}

	webRwMux.RLock()
	defer webRwMux.RUnlock()

	funcMap, ok := webRuleMap[resource]
	if !ok {
		return nil, false
	}
	behavior, ok := funcMap[functionType]
	if !ok {
		return nil, false
	}
	return behavior, true
}

func GetRpcFallbackBehavior(resource string, blockType base.BlockType) (*RpcBlockFallbackBehavior, bool) {
	functionType, ok := getFuncTypeByBlockType(blockType)
	if !ok {
		return nil, false
	}

	rpcRwMux.RLock()
	defer rpcRwMux.RUnlock()

	funcMap, ok := rpcRuleMap[resource]
	if !ok {
		return nil, false
	}
	behavior, ok := funcMap[functionType]
	if !ok {
		return nil, false
	}
	return behavior, true
}
