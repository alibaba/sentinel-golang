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

package server

import (
	"llm_token_ratelimit/llm_client"
	"net/http"
	"time"

	llmtokenratelimit "github.com/alibaba/sentinel-golang/core/llm_token_ratelimit"
	"github.com/gin-gonic/gin"
)

func (s *Server) chatCompletion(c *gin.Context) {
	infos, err := bindJSONFromCache[llm_client.LLMRequestInfos](c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "failed to process LLM request",
				"details": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}
	client, err := llm_client.NewLLMClient(infos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "failed to create LLM client",
				"details": err.Error(),
				"type":    "client_error",
			},
		})
		return
	}
	response, err := client.GenerateContent(infos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "failed to generate content",
				"details": err.Error(),
				"type":    "generation_error",
			},
		})
		return
	}

	if response == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "received nil response from LLM",
				"type":    "response_error",
			},
		})
		return
	}

	usedTokenInfos, err := llmtokenratelimit.OpenAITokenExtractor(response.Usage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "failed to extract token usage info",
				"details": err.Error(),
				"type":    "extraction_error",
			},
		})
		return
	}
	c.Set(llmtokenratelimit.KeyUsedTokenInfos, usedTokenInfos)

	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"timestamp": time.Now().Unix(),
		"choices":   response.Content,
		"usage": gin.H{
			"input_tokens":  usedTokenInfos.InputTokens,
			"output_tokens": usedTokenInfos.OutputTokens,
			"total_tokens":  usedTokenInfos.TotalTokens,
		},
	})
}
