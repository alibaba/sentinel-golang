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

package llm_token_ratelimit

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/base"
)

type Context struct {
	mu          sync.RWMutex
	userContext map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		userContext: make(map[string]interface{}),
	}
}

func (ctx *Context) Set(key string, value interface{}) {
	if ctx == nil {
		return
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.userContext == nil {
		ctx.userContext = make(map[string]interface{})
	}
	ctx.userContext[key] = value
}

func (ctx *Context) Get(key string) interface{} {
	if ctx == nil {
		return nil
	}

	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if ctx.userContext != nil {
		return ctx.userContext[key]
	}
	return nil
}

func (ctx *Context) extractArgs(entryCtx *base.EntryContext) {
	if ctx == nil || entryCtx == nil || entryCtx.Input == nil || entryCtx.Input.Args == nil {
		return
	}
	for _, arg := range entryCtx.Input.Args {
		if reqInfos, ok := arg.(*RequestInfos); ok && reqInfos != nil {
			ctx.Set(KeyRequestInfos, reqInfos)
		}
	}
}
