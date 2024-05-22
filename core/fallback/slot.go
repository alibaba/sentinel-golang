package fallback

func GetWebFallbackBehavior(resource string, functionType FunctionType) (*WebBlockFallbackBehavior, bool) {
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

func GetRpcFallbackBehavior(resource string, functionType FunctionType) (*RpcBlockFallbackBehavior, bool) {
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
