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

// Package config provides llm token rate limit mechanism.
//
// This package implements two core rate limiting strategies:
//  1. Fixed Window: A time-window based token limiting strategy.
//  2. PETA (Predictive Error Temporal Amortized): A predictive input tokens limiting strategy.
//
// Key features include:
//  1. Token-based rate limiting for LLM API calls.
//  2. Support for multiple token counting strategies.
//  3. Redis-backed distributed rate limiting to ensure consistency across clusters.
//  4. Token encoding support for estimating token usage of input content.
//  5. Adapters for seamless integration with common LLM frameworks.
//
// This functionality enables fine-grained control over LLM API usage based on token consumption, suitable for managing access to large language models with strict token limits.
package llm_token_ratelimit
