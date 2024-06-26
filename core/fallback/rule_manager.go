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
	"encoding/json"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"reflect"
	"sync"
)

var (
	webRuleMap      = make(map[string]map[FunctionType]*WebBlockFallbackBehavior)
	rpcRuleMap      = make(map[string]map[FunctionType]*RpcBlockFallbackBehavior)
	currentWebRules = make(map[string]map[FunctionType]*WebBlockFallbackBehavior)
	currentRpcRules = make(map[string]map[FunctionType]*RpcBlockFallbackBehavior)
	webRwMux        = &sync.RWMutex{}
	rpcRwMux        = &sync.RWMutex{}
	updateRuleMux   = &sync.RWMutex{}
)

func isValidWebFallbackBehavior(behavior *WebBlockFallbackBehavior) bool {
	if behavior == nil {
		return false
	}
	if behavior.WebRespContentType != 0 && behavior.WebRespContentType != 1 {
		return false
	}
	if behavior.WebRespStatusCode < 0 || behavior.WebRespStatusCode > 600 {
		return false
	}
	return true
}

func isValidRpcFallbackBehavior(behavior *RpcBlockFallbackBehavior) bool {
	if behavior == nil {
		return false
	}
	if behavior.RpcFallbackMode != 0 && behavior.RpcFallbackMode != 1 {
		return false
	}
	return true
}

func LoadRules(rules []*Rule) (bool, error) {
	resWebRuleMap := make(map[string]map[FunctionType]*WebBlockFallbackBehavior)
	resRpcRuleMap := make(map[string]map[FunctionType]*RpcBlockFallbackBehavior)
	for _, rule := range rules {
		b, err := json.Marshal(rule.FallbackBehavior)
		if err != nil {
			logging.Warn("[Fallback] marshal web fall back behavior failed", "reason", err.Error())
			continue
		}
		switch rule.TargetResourceType {
		case WebResourceType:
			var webBehavior *WebBlockFallbackBehavior
			err := json.Unmarshal(b, &webBehavior)
			if err != nil {
				logging.Warn("[Fallback] unmarshal web fall back behavior failed", "reason", err.Error())
				continue
			}
			if !isValidWebFallbackBehavior(webBehavior) {
				logging.Warn("[Fallback] invalid web fall back behavior", "behavior", webBehavior)
				continue
			}

			for resource, funcTypeList := range rule.TargetMap {
				if resource == "" || len(funcTypeList) == 0 {
					continue
				}
				var behaviorMap map[FunctionType]*WebBlockFallbackBehavior
				var ok bool
				if behaviorMap, ok = resWebRuleMap[resource]; !ok {
					behaviorMap = make(map[FunctionType]*WebBlockFallbackBehavior)
					resWebRuleMap[resource] = behaviorMap
				}

				for _, functionType := range funcTypeList {
					behaviorMap[functionType] = webBehavior
				}
			}
		case RpcResourceType:
			var rpcBehavior *RpcBlockFallbackBehavior
			err := json.Unmarshal(b, &rpcBehavior)
			if err != nil {
				logging.Warn("[Fallback] unmarshal rpc fall back behavior failed", "reason", err.Error())
				continue
			}
			if !isValidRpcFallbackBehavior(rpcBehavior) {
				logging.Warn("[Fallback] invalid rpc fall back behavior", "behavior", rpcBehavior)
				continue
			}

			for resource, funcTypeList := range rule.TargetMap {
				var behaviorMap map[FunctionType]*RpcBlockFallbackBehavior
				var ok bool
				if behaviorMap, ok = resRpcRuleMap[resource]; !ok {
					behaviorMap = make(map[FunctionType]*RpcBlockFallbackBehavior)
					resRpcRuleMap[resource] = behaviorMap
				}

				for _, functionType := range funcTypeList {
					behaviorMap[functionType] = rpcBehavior
				}
			}
		default:
			logging.Warn("[Fallback] unsupported resource type", "resourceType", rule.TargetResourceType)
			continue
		}
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	var err error
	var updated bool
	isEqual := reflect.DeepEqual(currentWebRules, resWebRuleMap)
	if !isEqual {
		updateErr := onWebRuleUpdate(resWebRuleMap)
		if updateErr != nil {
			logging.Error(updateErr, "[Fallback] update web rule failed")
			err = updateErr
		} else {
			updated = true
		}
	} else {
		logging.Info("[Fallback] Web load rules is the same with current rules, so ignore load operation.")
	}
	isEqual = reflect.DeepEqual(currentRpcRules, resRpcRuleMap)
	if !isEqual {
		updateErr := onRpcRuleUpdate(resRpcRuleMap)
		if updateErr != nil {
			logging.Error(updateErr, "[Fallback] update rpc rule failed")
			err = updateErr
		} else {
			updated = true
		}
	} else {
		logging.Info("[Fallback] Rpc load rules is the same with current rules, so ignore load operation.")
	}
	return updated, err
}

func onWebRuleUpdate(rawWebRuleMap map[string]map[FunctionType]*WebBlockFallbackBehavior) error {
	start := util.CurrentTimeNano()
	webRwMux.Lock()
	webRuleMap = rawWebRuleMap
	webRwMux.Unlock()
	currentWebRules = rawWebRuleMap
	logging.Debug("[Fallback onWebRuleUpdate] Time statistic(ns) for updating web fallback rule", "timeCost", util.CurrentTimeNano()-start)
	return nil
}

func onRpcRuleUpdate(rawRpcRuleMap map[string]map[FunctionType]*RpcBlockFallbackBehavior) error {
	start := util.CurrentTimeNano()
	rpcRwMux.Lock()
	rpcRuleMap = rawRpcRuleMap
	rpcRwMux.Unlock()
	currentRpcRules = rawRpcRuleMap
	logging.Debug("[Fallback onRpcRuleUpdate] Time statistic(ns) for updating rpc fallback rule", "timeCost", util.CurrentTimeNano()-start)
	return nil
}
