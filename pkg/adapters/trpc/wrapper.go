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
package trpc

import (
	"context"
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"trpc.group/trpc-go/trpc-agent-go/model"
)

type LLMWrapper struct {
	llm     model.Model
	options *options
}

func NewLLMWrapper(llm model.Model, opts ...Option) *LLMWrapper {
	return &LLMWrapper{
		llm:     llm,
		options: evaluateOptions(opts...),
	}
}

func (w *LLMWrapper) GenerateContent(ctx context.Context, messages []model.Message) (*model.Response, error) {
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

	// Get Request
	request := &model.Request{
		Messages: messages,
	}

	responseChan, llmErr := w.llm.GenerateContent(ctx, request)
	if llmErr != nil {
		return nil, llmErr
	}

	// receive response
	var finalResponse *model.Response
	for response := range responseChan {
		if response.Error != nil {
			return nil, fmt.Errorf("API error: %s", response.Error.Message)
		}
		finalResponse = response
	}

	if finalResponse == nil {
		return nil, fmt.Errorf("no response received from LLM")
	}

	if err := w.validateResponse(finalResponse); err != nil {
		return nil, err
	}

	usedTokenInfos := &llmtokenratelimit.UsedTokenInfos{}
	if w.options.usedTokenInfosExtract != nil {
		usedTokenInfos = w.options.usedTokenInfosExtract(finalResponse.Usage)
	} else {
		// fallback to OpenAI extractor
		tokenData := map[string]any{}
		if finalResponse.Usage != nil {
			tokenData["prompt_tokens"] = finalResponse.Usage.PromptTokens
			tokenData["completion_tokens"] = finalResponse.Usage.CompletionTokens
			tokenData["total_tokens"] = finalResponse.Usage.TotalTokens
		}
		infos, err := llmtokenratelimit.OpenAITokenExtractor(tokenData)
		if err != nil {
			return nil, err
		}
		usedTokenInfos = infos
	}

	entry.SetPair(llmtokenratelimit.KeyUsedTokenInfos, usedTokenInfos)
	entry.Exit()

	return finalResponse, nil
}

func (w *LLMWrapper) validateResponse(response *model.Response) error {
	if response == nil || len(response.Choices) == 0 {
		return fmt.Errorf("llm response is nil or empty")
	}

	// check Usage
	if response.Usage == nil {
		return fmt.Errorf("llm response missing Usage info")
	}

	// check token
	if response.Usage.PromptTokens == 0 {
		return fmt.Errorf("llm response has invalid Prompttoken usage info")
	}

	if response.Usage.CompletionTokens == 0 {
		return fmt.Errorf("llm response has invalid CompletionTokens usage info")
	}

	if response.Usage.TotalTokens == 0 {
		return fmt.Errorf("llm response has invalid TotalTokens usage info")
	}
	return nil
}
