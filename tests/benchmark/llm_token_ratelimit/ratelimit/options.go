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

package ratelimit

import (
	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"github.com/gin-gonic/gin"
)

type Option func(*options)

type options struct {
	resourceExtract     func(*gin.Context) string
	blockFallback       func(*gin.Context)
	requestInfosExtract func(*gin.Context) *llmtokenratelimit.RequestInfos
	promptsExtract      func(*gin.Context) []string
}

func WithBlockFallback(fn func(ctx *gin.Context)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}

func WithResourceExtractor(fn func(*gin.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

func WithRequestInfosExtractor(fn func(*gin.Context) *llmtokenratelimit.RequestInfos) Option {
	return func(opts *options) {
		opts.requestInfosExtract = fn
	}
}

func WithPromptsExtractor(fn func(*gin.Context) []string) Option {
	return func(opts *options) {
		opts.promptsExtract = fn
	}
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}

	return optCopy
}
