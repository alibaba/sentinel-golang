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

package llm_client

import (
	"context"
	"fmt"
	"os"

	langchaingo_llms "github.com/tmc/langchaingo/llms"
	langchaingo_openai "github.com/tmc/langchaingo/llms/openai"

	eino_openai "github.com/cloudwego/eino-ext/components/model/openai"
	eino_model "github.com/cloudwego/eino/components/model"
	eino_schema "github.com/cloudwego/eino/schema"
)

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequestInfos struct {
	Provider LLMProvider  `json:"provider"`
	Messages []LLMMessage `json:"messages"`
	Model    string       `json:"model"`
}

type LLMResponseInfos struct {
	Content string         `json:"content"`
	Usage   map[string]any `json:"usage,omitempty"`
}

type LLMProvider int32

const (
	LangChain LLMProvider = iota
	Eino
)

func (p LLMProvider) String() string {
	switch p {
	case LangChain:
		return "langchain"
	case Eino:
		return "eino"
	default:
		return "unknown"
	}
}

func ParseLLMProvider(s string) (LLMProvider, error) {
	switch s {
	case "langchain":
		return LangChain, nil
	case "eino":
		return Eino, nil
	default:
		return 0, fmt.Errorf("unknown LLM provider: %s", s)
	}
}

func (p LLMProvider) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *LLMProvider) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	provider, err := ParseLLMProvider(s)
	if err != nil {
		return err
	}
	*p = provider
	return nil
}

// ================================= LLMClient ====================================

type LLMClient interface {
	GenerateContent(infos *LLMRequestInfos) (*LLMResponseInfos, error)
	GetProvider() LLMProvider
}

func NewLLMClient(infos *LLMRequestInfos) (LLMClient, error) {
	switch infos.Provider {
	case LangChain:
		return NewLangChainClient(infos.Model)
	case Eino:
		return NewEinoClient(infos.Model)
	default:
		return nil, fmt.Errorf("unsupported provider: %v", infos.Provider)
	}
}

// ================================= LangChainClient ================================

type LangChainClient struct {
	llm langchaingo_llms.Model
}

func NewLangChainClient(model string) (LLMClient, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY environment variable is not set")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("LLM_BASE_URL environment variable is not set")
	}

	llm, err := langchaingo_openai.New(
		langchaingo_openai.WithToken(apiKey),
		langchaingo_openai.WithBaseURL(baseURL),
		langchaingo_openai.WithModel(model),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChain LLM client: %w", err)
	}

	return &LangChainClient{
		llm: llm,
	}, nil
}

func (c *LangChainClient) GenerateContent(infos *LLMRequestInfos) (*LLMResponseInfos, error) {
	if infos == nil || infos.Messages == nil {
		return nil, fmt.Errorf("invalid request infos")
	}
	content := make([]langchaingo_llms.MessageContent, len(infos.Messages))
	for i, msg := range infos.Messages {
		content[i] = langchaingo_llms.TextParts(langchaingo_llms.ChatMessageType(msg.Role), msg.Content)
	}
	completion, err := c.llm.GenerateContent(context.Background(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}
	return &LLMResponseInfos{
		Content: completion.Choices[0].Content,
		Usage: map[string]any{
			"prompt_tokens":     completion.Choices[0].GenerationInfo["PromptTokens"],
			"completion_tokens": completion.Choices[0].GenerationInfo["CompletionTokens"],
			"total_tokens":      completion.Choices[0].GenerationInfo["TotalTokens"],
		},
	}, nil
}

func (c *LangChainClient) GetProvider() LLMProvider {
	return LangChain
}

// ================================= EinoClient ====================================

type EinoClient struct {
	llm eino_model.BaseChatModel
}

func NewEinoClient(model string) (LLMClient, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY environment variable is not set")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("LLM_BASE_URL environment variable is not set")
	}

	llm, err := eino_openai.NewChatModel(context.Background(), &eino_openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Eino LLM client: %w", err)
	}

	return &EinoClient{
		llm: llm,
	}, nil
}

func (c *EinoClient) GenerateContent(infos *LLMRequestInfos) (*LLMResponseInfos, error) {
	if infos == nil || infos.Messages == nil {
		return nil, fmt.Errorf("invalid request infos")
	}
	content := make([]*eino_schema.Message, len(infos.Messages))
	for i, msg := range infos.Messages {
		content[i] = &eino_schema.Message{
			Role:    eino_schema.RoleType(msg.Role),
			Content: msg.Content,
		}
	}
	completion, err := c.llm.Generate(context.Background(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}
	return &LLMResponseInfos{
		Content: completion.Content,
		Usage: map[string]any{
			"prompt_tokens":     completion.ResponseMeta.Usage.PromptTokens,
			"completion_tokens": completion.ResponseMeta.Usage.CompletionTokens,
			"total_tokens":      completion.ResponseMeta.Usage.TotalTokens,
		},
	}, nil
}

func (c *EinoClient) GetProvider() LLMProvider {
	return Eino
}
