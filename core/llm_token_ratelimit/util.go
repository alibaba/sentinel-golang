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
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"unsafe"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/spaolacci/murmur3"
)

func generateHash(parts ...string) string {
	h := murmur3.New64()
	for i, part := range parts {
		if i > 0 {
			h.Write([]byte{0}) // separator
		}
		h.Write([]byte(part))
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func parseRedisResponse(ctx *Context, response interface{}) []int64 {
	if response == nil {
		return nil
	}

	resultSlice, ok := response.([]interface{})
	if !ok || resultSlice == nil {
		return nil
	}

	result := make([]int64, len(resultSlice))
	for i, v := range resultSlice {
		switch val := v.(type) {
		case int64:
			result[i] = val
		case string:
			num, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				logging.Error(err, "failed to parse redis response element in llm_token_ratelimit.parseRedisResponse()",
					"index", i,
					"value", val,
					"error", err.Error(),
					"requestID", ctx.Get(KeyRequestID),
				)
				return nil
			}
			result[i] = num
		case int:
			result[i] = int64(val)
		case float64:
			result[i] = int64(val)
		default:
			logging.Error(errors.New("unknown error"), "unexpected redis response element type in llm_token_ratelimit.parseRedisResponse()",
				"index", i,
				"value", v,
				"type", fmt.Sprintf("%T", v),
				"requestID", ctx.Get(KeyRequestID),
			)
			return nil
		}
	}
	return result
}

func generateRandomString(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)

	for i, cache, remain := n-1, rand.Int63(), RandomLetterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), RandomLetterIdxMax
		}
		if idx := int(cache & RandomLetterIdxMask); idx < len(RandomLetterBytes) {
			b[i] = RandomLetterBytes[idx]
			i--
		}
		cache >>= RandomLetterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func generateUUID() string {
	return uuid.NewString()
}

func deepCopyByCopier(src, dest interface{}) error {
	if src == nil {
		return errors.New("src is nil")
	}
	if dest == nil {
		return errors.New("dest is nil")
	}
	return copier.CopyWithOption(dest, src, copier.Option{
		DeepCopy: true,
	})
}
