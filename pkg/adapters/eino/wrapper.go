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
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type LLMWrapper struct {
	llm     model.BaseChatModel
	options *options
}

func NewLLMWrapper(llm model.BaseChatModel, opts ...Option) *LLMWrapper {
	return &LLMWrapper{
		llm:     llm,
		options: evaluateOptions(opts...),
	}
}

func (w *LLMWrapper) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	resource := w.options.defaultResource
	if w.options.resourceExtract != nil {
		resource = w.options.resourceExtract(ctx)
	}

	prompts := []string{}
	if w.options.promptsExtract != nil {
		prompts = w.options.promptsExtract(input)
	}

	reqInfos := llmtokenratelimit.GenerateRequestInfos(
		llmtokenratelimit.WithPrompts(prompts),
	)
	if w.options.requestInfosExtract != nil {
		reqInfos = w.options.requestInfosExtract(ctx)
	}

	// Check
	entry, err := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(reqInfos))

	if err != nil {
		// Block
		if w.options.blockFallback != nil {
			w.options.blockFallback(ctx)
		}
		return nil, err
	}
	// Pass
	response, llmErr := w.llm.Generate(ctx, input, opts...)
	if llmErr != nil {
		return nil, llmErr
	}
	if err := w.validateResponse(response); err != nil {
		return nil, err
	}

	usedTokenInfos := &llmtokenratelimit.UsedTokenInfos{}
	if w.options.usedTokenInfosExtract != nil {
		usedTokenInfos = w.options.usedTokenInfosExtract(response.ResponseMeta.Usage)
	} else {
		// fallback to OpenAI extractor
		infos, err := llmtokenratelimit.OpenAITokenExtractor(map[string]any{
			"prompt_tokens":     response.ResponseMeta.Usage.PromptTokens,
			"completion_tokens": response.ResponseMeta.Usage.CompletionTokens,
			"total_tokens":      response.ResponseMeta.Usage.TotalTokens,
		})
		if err != nil {
			return nil, err
		}
		usedTokenInfos = infos
	}

	entry.SetPair(llmtokenratelimit.KeyUsedTokenInfos, usedTokenInfos)
	entry.Exit()

	return response, nil
}

func (w *LLMWrapper) validateResponse(response *schema.Message) error {
	if response == nil || response.ResponseMeta == nil || response.ResponseMeta.Usage == nil {
		return fmt.Errorf("llm response is nil or empty or missing Usage infos")
	}
	return nil
}
