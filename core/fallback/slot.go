// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
