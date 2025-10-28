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

// ================================= Config ====================================
const (
	DefaultResourcePattern        string = ".*"
	DefaultIdentifierValuePattern string = ".*"
	DefaultKeyPattern             string = ".*"

	DefaultRedisAddrName     string = "127.0.0.1"
	DefaultRedisAddrPort     int32  = 6379
	DefaultRedisTimeout      int32  = 0 // milliseconds
	DefaultRedisPoolSize     int32  = 10
	DefaultRedisMinIdleConns int32  = 5
	DefaultRedisMaxRetries   int32  = 3

	DefaultErrorCode    int32  = 429
	DefaultErrorMessage string = "Too Many Requests"
)

var DefaultTokenEncodingModel = map[TokenEncoderProvider]string{
	OpenAIEncoderProvider: "gpt-4",
}

// ================================= CommonError ==============================
const (
	ErrorTimeDuration int64 = -1
)

// ================================= Context ==================================
const (
	KeyContext         string = "SentinelLLMTokenRatelimitContext"
	KeyRequestInfos    string = "SentinelLLMTokenRatelimitReqInfos"
	KeyUsedTokenInfos  string = "SentinelLLMTokenRatelimitUsedTokenInfos"
	KeyMatchedRules    string = "SentinelLLMTokenRatelimitMatchedRules"
	KeyResponseHeaders string = "SentinelLLMTokenRatelimitResponseHeaders"
	KeyRequestID       string = "SentinelLLMTokenRatelimitRequestID"
	KeyErrorCode       string = "SentinelLLMTokenRatelimitErrorCode"
	KeyErrorMessage    string = "SentinelLLMTokenRatelimitErrorMessage"
)

// ================================= RedisRatelimitKeyFormat ==================
const (
	RedisRatelimitKeyFormat string = "sentinel-go:llm-token-ratelimit:resource-%s:%s:%s:%d:%s" // hashedResource, strategy, identifierType, timeWindow, tokenCountStrategy
)

// ================================= ResponseHeader ==================
const (
	ResponseHeaderRequestID       string = "X-Sentinel-LLM-Token-Ratelimit-RequestID"
	ResponseHeaderRemainingTokens string = "X-Sentinel-LLM-Token-Ratelimit-RemainingTokens"
	ResponseHeaderWaitingTime     string = "X-Sentinel-LLM-Token-Ratelimit-WaitingTime"
)

// ================================= FixedWindowStrategy ======================

// ================================= PETAStrategy =============================
const (
	PETANoWaiting                 int64  = 0
	PETACorrectOK                 int64  = 0
	PETACorrectUnderestimateError int64  = 1
	PETACorrectOverestimateError  int64  = 2
	PETASlidingWindowKeyFormat    string = "{shard-%s}:sliding-window:%s" // hashTag, redisRatelimitKey
	PETATokenBucketKeyFormat      string = "{shard-%s}:token-bucket:%s"   // hashTag, redisRatelimitKey
	PETARandomStringLength        int    = 16
)

// ================================= Generate Random String ===================
const (
	RandomLetterBytes   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	RandomLetterIdxBits = 6
	RandomLetterIdxMask = 1<<RandomLetterIdxBits - 1
	RandomLetterIdxMax  = 63 / RandomLetterIdxBits
)

// ================================= TokenEncoder =============================
const (
	TokenEncoderKeyFormat string = "{shard-%s}:token-encoder:%s:%s:%s" // hashTag, provider, model, redisRatelimitKey
)

// ================================= MetricLogger =============================
const (
	MetricFileNameSuffix = "llm-token-ratelimit-metrics.log"
	DefaultMaxFileSize   = 50 * 1024 * 1024
	DefaultMaxFileAmount = 7
	DefaultBufferSize    = 1000
	DefaultFlushInterval = 1 // second
)
