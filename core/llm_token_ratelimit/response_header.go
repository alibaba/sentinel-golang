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

type ResponseHeader struct {
	headers      map[string]string
	ErrorCode    int32
	ErrorMessage string
}

func NewResponseHeader() *ResponseHeader {
	return &ResponseHeader{
		headers: make(map[string]string),
	}
}

func (rh *ResponseHeader) Set(key, value string) {
	if rh == nil {
		return
	}

	if rh.headers == nil {
		rh.headers = make(map[string]string)
	}
	rh.headers[key] = value
}

func (rh *ResponseHeader) Get(key string) string {
	if rh == nil {
		return ""
	}

	if rh.headers == nil {
		return ""
	}
	return rh.headers[key]
}

func (rh *ResponseHeader) GetAll() map[string]string {
	if rh == nil {
		return nil
	}

	return rh.headers
}
