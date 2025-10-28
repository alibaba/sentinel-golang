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
	"github.com/alibaba/sentinel-golang/util"
)

type AllIdentifierChecker struct{}

func (f *AllIdentifierChecker) Check(ctx *Context, infos *RequestInfos, identifier Identifier, pattern string) bool {
	if f == nil {
		return false
	}
	if globalRuleMatcher == nil {
		return true // system not ready, allow all
	}
	if infos == nil {
		return true // allow nil for global rate limit
	}
	for identifierType, checker := range globalRuleMatcher.IdentifierCheckers {
		if identifierType == AllIdentifier {
			continue
		}
		if checker.Check(ctx, infos, identifier, pattern) {
			return true
		}
	}
	return false
}

type HeaderChecker struct{}

func (f *HeaderChecker) Check(ctx *Context, infos *RequestInfos, identifier Identifier, pattern string) bool {
	if f == nil {
		return false
	}
	if infos == nil || infos.Headers == nil {
		return true // allow nil for global rate limit
	}
	for key, values := range infos.Headers {
		if len(values) == 0 {
			values = []string{pattern}
		}
		for _, v := range values {
			if util.RegexMatch(identifier.Value, key) && util.RegexMatch(pattern, v) {
				return true
			}
		}
	}
	return false
}
