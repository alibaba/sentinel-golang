// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package langchaingo

import (
	"context"
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"github.com/tmc/langchaingo/llms"
)

type LLMWrapper struct {
	llm     llms.Model
	options *options
}

func NewLLMWrapper(llm llms.Model, opts ...Option) *LLMWrapper {
	return &LLMWrapper{
		llm:     llm,
		options: evaluateOptions(opts...),
	}
}

func (w *LLMWrapper) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	resource := w.options.defaultResource
	if w.options.resourceExtract != nil {
		resource = w.options.resourceExtract(ctx)
	}

	prompts := []string{}
	if w.options.promptsExtract != nil {
		prompts = w.options.promptsExtract(messages)
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
	response, llmErr := w.llm.GenerateContent(ctx, messages, options...)
	if llmErr != nil {
		return nil, llmErr
	}
	if err := w.validateResponse(response); err != nil {
		return nil, err
	}

	usedTokenInfos := &llmtokenratelimit.UsedTokenInfos{}
	if w.options.usedTokenInfosExtract != nil {
		usedTokenInfos = w.options.usedTokenInfosExtract(response.Choices[0].GenerationInfo)
	} else {
		// fallback to OpenAI extractor
		infos, err := llmtokenratelimit.OpenAITokenExtractor(map[string]any{
			"prompt_tokens":     response.Choices[0].GenerationInfo["PromptTokens"],
			"completion_tokens": response.Choices[0].GenerationInfo["CompletionTokens"],
			"total_tokens":      response.Choices[0].GenerationInfo["TotalTokens"],
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

func (w *LLMWrapper) validateResponse(response *llms.ContentResponse) error {
	if response == nil || len(response.Choices) == 0 || response.Choices[0].GenerationInfo == nil {
		return fmt.Errorf("llm response is nil or empty or missing GenerationInfo")
	}
	_, exists := response.Choices[0].GenerationInfo["PromptTokens"]
	if !exists {
		return fmt.Errorf("llm response missing PromptTokens info")
	}
	_, exists = response.Choices[0].GenerationInfo["CompletionTokens"]
	if !exists {
		return fmt.Errorf("llm response missing CompletionTokens info")
	}
	_, exists = response.Choices[0].GenerationInfo["TotalTokens"]
	if !exists {
		return fmt.Errorf("llm response missing TotalTokens info")
	}
	return nil
}
