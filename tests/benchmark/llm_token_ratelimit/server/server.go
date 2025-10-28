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
	"bytes"
	"fmt"
	"io"
	"llm_token_ratelimit/llm_client"
	"llm_token_ratelimit/ratelimit"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	KeyRawBody = "rawBody"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func cacheBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to read request body",
			})
			c.Abort()
			return
		}

		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Set(KeyRawBody, bodyBytes)

		c.Next()
	}
}

func bindJSONFromCache[T any](c *gin.Context) (*T, error) {
	rawBody, exists := c.Get(KeyRawBody)
	if !exists {
		return nil, fmt.Errorf("raw body not found in context")
	}

	bodyBytes, ok := rawBody.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid raw body type")
	}

	tempBody := io.NopCloser(bytes.NewBuffer(bodyBytes))
	originalBody := c.Request.Body
	c.Request.Body = tempBody

	var result T
	err := c.ShouldBindJSON(&result)

	c.Request.Body = originalBody

	if err != nil {
		return nil, err
	}

	return &result, nil
}

type Server struct {
	engine *gin.Engine
	ip     string
	port   uint16
}

func NewServer(ip string, port uint16) *Server {
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(cacheBodyMiddleware())
	engine.Use(ratelimit.SentinelMiddleware(
		ratelimit.WithPromptsExtractor(func(c *gin.Context) []string {
			infos, err := bindJSONFromCache[llm_client.LLMRequestInfos](c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"message": "failed to process LLM request",
						"details": err.Error(),
						"type":    "api_error",
					},
				})
				return nil
			}
			prompts := make([]string, 0, len(infos.Messages))
			for _, msg := range infos.Messages {
				prompts = append(prompts, msg.Content)
			}
			return prompts
		}),
	))

	return &Server{
		engine: engine,
		ip:     ip,
		port:   port,
	}
}

func (s *Server) setupRoutes() {
	s.engine.POST("/v1/chat/completion/fixed_window", s.chatCompletion)
	s.engine.POST("/v1/chat/completion/peta", s.chatCompletion)
}

func (s *Server) Start() error {
	s.setupRoutes()
	addr := fmt.Sprintf("%s:%d", s.ip, s.port)
	return s.engine.Run(addr)
}

func StartServer(ip string, port uint16) {
	server := NewServer(ip, port)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
