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
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkoukk/tiktoken-go"
)

// ================================= TokenEncoder ====================================
var (
	tokenEncoderMap      = make(map[TokenEncoding]TokenEncoder)
	tokenEncoderMapRWMux = &sync.RWMutex{}
)

type TokenEncoder interface {
	CountTokens(ctx *Context, prompts []string, rule *MatchedRule) (int, error)
}

func NewTokenEncoder(ctx *Context, encoding TokenEncoding) TokenEncoder {
	var encoder TokenEncoder
	switch encoding.Provider {
	case OpenAIEncoderProvider:
		encoder = NewOpenAIEncoder(ctx, encoding)
	default:
		logging.Warn("[LLMTokenRateLimit] unsupported token encoder provider, falling back to OpenAIEncoder",
			"unsupported encoder provider", encoding.Provider,
			"requestID", ctx.Get(KeyRequestID),
		)
		encoder = NewOpenAIEncoder(ctx, encoding) // Fallback to OpenAIEncoder for unsupported providers
	}
	tokenEncoderMapRWMux.Lock()
	defer tokenEncoderMapRWMux.Unlock()
	tokenEncoderMap[encoding] = encoder
	return encoder
}

func LookupTokenEncoder(ctx *Context, encoding TokenEncoding) TokenEncoder {
	tokenEncoderMapRWMux.RLock()
	defer tokenEncoderMapRWMux.RUnlock()
	return tokenEncoderMap[encoding]
}

// ================================= OpenAIEncoder ====================================

type OpenAIEncoder struct {
	Model   string
	Encoder *tiktoken.Tiktoken
}

func NewOpenAIEncoder(ctx *Context, encoding TokenEncoding) *OpenAIEncoder {
	encoder, err := tiktoken.EncodingForModel(encoding.Model)
	actualModel := encoding.Model

	if err != nil {
		actualModel = DefaultTokenEncodingModel[OpenAIEncoderProvider]
		logging.Warn("[LLMTokenRateLimit] model not supported, falling back to default model",
			"unsupported model", encoding.Model,
			"default model", actualModel,
			"requestID", ctx.Get(KeyRequestID),
		)
		encoder, _ = tiktoken.EncodingForModel(actualModel)
	}

	return &OpenAIEncoder{
		Model:   actualModel,
		Encoder: encoder,
	}
}

func (e *OpenAIEncoder) CountTokens(ctx *Context, prompts []string, rule *MatchedRule) (int, error) {
	if e == nil {
		return 0, fmt.Errorf("OpenAIEncoder is nil")
	}
	if e.Encoder == nil {
		return 0, fmt.Errorf("OpenAIEncoder's encoder is nil for model: %s", e.Model)
	}
	if len(prompts) == 0 {
		return 0, nil // No prompts to count tokens
	}
	// Concatenate prompts
	var builder strings.Builder
	for _, prompt := range prompts {
		builder.WriteString(prompt)
	}
	token := e.Encoder.Encode(builder.String(), nil, nil)
	return len(token), nil
}
