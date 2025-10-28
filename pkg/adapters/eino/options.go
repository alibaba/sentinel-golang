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

package eino

import (
	"context"

	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"github.com/cloudwego/eino/schema"
)

type Option func(*options)

type options struct {
	defaultResource       string
	resourceExtract       func(context.Context) string
	blockFallback         func(context.Context)
	requestInfosExtract   func(context.Context) *llmtokenratelimit.RequestInfos
	promptsExtract        func([]*schema.Message) []string
	usedTokenInfosExtract func(interface{}) *llmtokenratelimit.UsedTokenInfos
}

func WithDefaultResource(resource string) Option {
	return func(o *options) {
		o.defaultResource = resource
	}
}

func WithResourceExtract(fn func(context.Context) string) Option {
	return func(o *options) {
		o.resourceExtract = fn
	}
}

func WithBlockFallback(fn func(context.Context)) Option {
	return func(o *options) {
		o.blockFallback = fn
	}
}

func WithRequestInfosExtract(fn func(context.Context) *llmtokenratelimit.RequestInfos) Option {
	return func(o *options) {
		o.requestInfosExtract = fn
	}
}

func WithPromptsExtract(fn func([]*schema.Message) []string) Option {
	return func(o *options) {
		o.promptsExtract = fn
	}
}

func WithUsedTokenInfosExtract(fn func(interface{}) *llmtokenratelimit.UsedTokenInfos) Option {
	return func(o *options) {
		o.usedTokenInfosExtract = fn
	}
}

func evaluateOptions(opts ...Option) *options {
	optCopy := &options{
		defaultResource: llmtokenratelimit.DefaultResourcePattern,
		promptsExtract: func(messages []*schema.Message) []string {
			prompts := make([]string, 0, len(messages))
			for _, msg := range messages {
				for _, part := range msg.MultiContent {
					prompts = append(prompts, part.Text)
				}
				if len(msg.Content) != 0 {
					prompts = append(prompts, msg.Content)
				}
			}
			return prompts
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(optCopy)
		}
	}
	return optCopy
}
