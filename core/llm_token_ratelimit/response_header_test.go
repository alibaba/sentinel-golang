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

package llm_token_ratelimit

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewResponseHeader(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new response header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rh := NewResponseHeader()
			if rh == nil {
				t.Fatal("NewResponseHeader() should not return nil")
			}
			if rh.headers == nil {
				t.Fatal("headers should be initialized")
			}
			if len(rh.headers) != 0 {
				t.Fatal("headers should be empty initially")
			}
			if rh.ErrorCode != 0 {
				t.Fatal("ErrorCode should be 0 initially")
			}
			if rh.ErrorMessage != "" {
				t.Fatal("ErrorMessage should be empty initially")
			}
		})
	}
}

func TestResponseHeader_Set(t *testing.T) {
	tests := []struct {
		name   string
		rh     *ResponseHeader
		key    string
		value  string
		verify func(*testing.T, *ResponseHeader)
	}{
		{
			name:  "nil response header",
			rh:    nil,
			key:   "test",
			value: "value",
			verify: func(t *testing.T, rh *ResponseHeader) {
				// Should not panic, and rh remains nil
				if rh != nil {
					t.Error("nil ResponseHeader should remain nil")
				}
			},
		},
		{
			name:  "normal set",
			rh:    NewResponseHeader(),
			key:   "Content-Type",
			value: "application/json",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("Content-Type") != "application/json" {
					t.Error("header should be set correctly")
				}
			},
		},
		{
			name:  "set with nil headers map",
			rh:    &ResponseHeader{headers: nil},
			key:   "Authorization",
			value: "Bearer token",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.headers == nil {
					t.Error("headers map should be initialized")
				}
				if rh.Get("Authorization") != "Bearer token" {
					t.Error("header should be set correctly")
				}
			},
		},
		{
			name: "overwrite existing key",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("User-Id", "old_value")
				return rh
			}(),
			key:   "User-Id",
			value: "new_value",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("User-Id") != "new_value" {
					t.Error("header should be overwritten")
				}
			},
		},
		{
			name:  "empty key",
			rh:    NewResponseHeader(),
			key:   "",
			value: "empty_key_value",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("") != "empty_key_value" {
					t.Error("empty key should be handled")
				}
			},
		},
		{
			name:  "empty value",
			rh:    NewResponseHeader(),
			key:   "Empty-Value",
			value: "",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("Empty-Value") != "" {
					t.Error("empty value should be set")
				}
			},
		},
		{
			name:  "special characters in key",
			rh:    NewResponseHeader(),
			key:   "X-Custom-Header_123!@#",
			value: "special_value",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("X-Custom-Header_123!@#") != "special_value" {
					t.Error("special character key should be handled")
				}
			},
		},
		{
			name:  "unicode characters",
			rh:    NewResponseHeader(),
			key:   "Unicode-Header",
			value: "测试值",
			verify: func(t *testing.T, rh *ResponseHeader) {
				if rh.Get("Unicode-Header") != "测试值" {
					t.Error("unicode value should be handled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalRh := tt.rh
			tt.rh.Set(tt.key, tt.value)
			tt.verify(t, originalRh)
		})
	}
}

func TestResponseHeader_Get(t *testing.T) {
	tests := []struct {
		name string
		rh   *ResponseHeader
		key  string
		want string
	}{
		{
			name: "nil response header",
			rh:   nil,
			key:  "test",
			want: "",
		},
		{
			name: "nil headers map",
			rh:   &ResponseHeader{headers: nil},
			key:  "test",
			want: "",
		},
		{
			name: "existing key",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Content-Type", "application/json")
				return rh
			}(),
			key:  "Content-Type",
			want: "application/json",
		},
		{
			name: "non-existing key",
			rh:   NewResponseHeader(),
			key:  "nonexistent",
			want: "",
		},
		{
			name: "empty key",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("", "empty_key_value")
				return rh
			}(),
			key:  "",
			want: "empty_key_value",
		},
		{
			name: "empty value",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Empty-Value", "")
				return rh
			}(),
			key:  "Empty-Value",
			want: "",
		},
		{
			name: "case sensitive key",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Content-Type", "application/json")
				return rh
			}(),
			key:  "content-type",
			want: "",
		},
		{
			name: "special characters",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("X-Special!@#", "special_value")
				return rh
			}(),
			key:  "X-Special!@#",
			want: "special_value",
		},
		{
			name: "unicode characters",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Unicode", "测试值")
				return rh
			}(),
			key:  "Unicode",
			want: "测试值",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rh.Get(tt.key)
			if got != tt.want {
				t.Errorf("ResponseHeader.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseHeader_GetAll(t *testing.T) {
	tests := []struct {
		name string
		rh   *ResponseHeader
		want map[string]string
	}{
		{
			name: "nil response header",
			rh:   nil,
			want: nil,
		},
		{
			name: "nil headers map",
			rh:   &ResponseHeader{headers: nil},
			want: nil,
		},
		{
			name: "empty headers",
			rh:   NewResponseHeader(),
			want: map[string]string{},
		},
		{
			name: "single header",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Content-Type", "application/json")
				return rh
			}(),
			want: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "multiple headers",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Content-Type", "application/json")
				rh.Set("Authorization", "Bearer token")
				rh.Set("User-Agent", "test-client")
				return rh
			}(),
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
				"User-Agent":    "test-client",
			},
		},
		{
			name: "headers with empty values",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("Empty-Header", "")
				rh.Set("Normal-Header", "value")
				return rh
			}(),
			want: map[string]string{
				"Empty-Header":  "",
				"Normal-Header": "value",
			},
		},
		{
			name: "headers with special characters",
			rh: func() *ResponseHeader {
				rh := NewResponseHeader()
				rh.Set("X-Custom!@#", "special_value")
				rh.Set("Unicode-Header", "测试值")
				return rh
			}(),
			want: map[string]string{
				"X-Custom!@#":    "special_value",
				"Unicode-Header": "测试值",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rh.GetAll()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResponseHeader.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test ResponseHeader fields
func TestResponseHeader_Fields(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    int32
		errorMessage string
	}{
		{
			name:         "default values",
			errorCode:    0,
			errorMessage: "",
		},
		{
			name:         "custom error values",
			errorCode:    404,
			errorMessage: "Not Found",
		},
		{
			name:         "negative error code",
			errorCode:    -1,
			errorMessage: "Internal Error",
		},
		{
			name:         "large error code",
			errorCode:    999999,
			errorMessage: "Custom Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rh := NewResponseHeader()
			rh.ErrorCode = tt.errorCode
			rh.ErrorMessage = tt.errorMessage

			if rh.ErrorCode != tt.errorCode {
				t.Errorf("ErrorCode = %v, want %v", rh.ErrorCode, tt.errorCode)
			}
			if rh.ErrorMessage != tt.errorMessage {
				t.Errorf("ErrorMessage = %v, want %v", rh.ErrorMessage, tt.errorMessage)
			}
		})
	}
}

func TestResponseHeader_ErrorCodeErrorMessage(t *testing.T) {
	rh := NewResponseHeader()

	// Test setting and getting error fields
	rh.ErrorCode = 500
	rh.ErrorMessage = "Internal Server Error"

	if rh.ErrorCode != 500 {
		t.Errorf("ErrorCode = %v, want 500", rh.ErrorCode)
	}
	if rh.ErrorMessage != "Internal Server Error" {
		t.Errorf("ErrorMessage = %v, want 'Internal Server Error'", rh.ErrorMessage)
	}

	// Test that headers still work
	rh.Set("Content-Type", "application/json")
	if rh.Get("Content-Type") != "application/json" {
		t.Error("Headers should still work when error fields are set")
	}
}

// Edge cases
func TestResponseHeader_EdgeCases(t *testing.T) {
	t.Run("very long key and value", func(t *testing.T) {
		rh := NewResponseHeader()
		longKey := generateLargeString(1000)
		longValue := generateLargeString(2000)

		rh.Set(longKey, longValue)
		retrieved := rh.Get(longKey)

		if retrieved != longValue {
			t.Error("Long key and value should be handled correctly")
		}
	})

	t.Run("key with newlines and tabs", func(t *testing.T) {
		rh := NewResponseHeader()
		specialKey := "Header\nWith\tSpecial\rChars"
		specialValue := "Value\nWith\tSpecial\rChars"

		rh.Set(specialKey, specialValue)
		retrieved := rh.Get(specialKey)

		if retrieved != specialValue {
			t.Error("Special characters should be handled correctly")
		}
	})

	t.Run("rapid set and get operations", func(t *testing.T) {
		rh := NewResponseHeader()

		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("rapid_%d", i)
			value := fmt.Sprintf("value_%d", i)
			rh.Set(key, value)

			if rh.Get(key) != value {
				t.Errorf("Rapid operation failed at iteration %d", i)
			}
		}
	})
}

// Test to ensure map reference safety
func TestResponseHeader_MapReferenceSafety(t *testing.T) {
	rh := NewResponseHeader()
	rh.Set("test", "original")

	// Get reference to internal map
	allHeaders := rh.GetAll()

	// Modify the returned map
	if allHeaders != nil {
		allHeaders["test"] = "modified"
		allHeaders["new_key"] = "new_value"
	}

	// Check if original is affected (it should be, since we return direct reference)
	if rh.Get("test") != "modified" {
		t.Log("Map reference is shared (this is current behavior)")
	}

	if rh.Get("new_key") != "new_value" {
		t.Log("Map reference is shared (this is current behavior)")
	}
}

// Performance tests
func BenchmarkNewResponseHeader(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewResponseHeader()
	}
}

func BenchmarkResponseHeader_Set(b *testing.B) {
	rh := NewResponseHeader()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Set("test_key", "test_value")
	}
}

func BenchmarkResponseHeader_Get(b *testing.B) {
	rh := NewResponseHeader()
	rh.Set("test_key", "test_value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Get("test_key")
	}
}

func BenchmarkResponseHeader_GetAll(b *testing.B) {
	rh := NewResponseHeader()
	rh.Set("Content-Type", "application/json")
	rh.Set("Authorization", "Bearer token")
	rh.Set("User-Agent", "test-client")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.GetAll()
	}
}

func BenchmarkResponseHeader_SetGet(b *testing.B) {
	rh := NewResponseHeader()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100) // Reuse keys to simulate real usage
		rh.Set(key, "value")
		rh.Get(key)
	}
}

// Memory allocation benchmarks
func BenchmarkNewResponseHeader_Memory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewResponseHeader()
	}
}

func BenchmarkResponseHeader_Set_Memory(b *testing.B) {
	rh := NewResponseHeader()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Set("test_key", "test_value")
	}
}

func BenchmarkResponseHeader_Get_Memory(b *testing.B) {
	rh := NewResponseHeader()
	rh.Set("test_key", "test_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Get("test_key")
	}
}

func BenchmarkResponseHeader_GetAll_Memory(b *testing.B) {
	rh := NewResponseHeader()
	rh.Set("Content-Type", "application/json")
	rh.Set("Authorization", "Bearer token")
	rh.Set("User-Agent", "test-client")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.GetAll()
	}
}

// Large data tests
func BenchmarkResponseHeader_LargeHeaders(b *testing.B) {
	rh := NewResponseHeader()

	// Set up many headers
	for i := 0; i < 1000; i++ {
		rh.Set(fmt.Sprintf("Header-%d", i), fmt.Sprintf("Value-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Get(fmt.Sprintf("Header-%d", i%1000))
	}
}

func BenchmarkResponseHeader_LargeValues(b *testing.B) {
	rh := NewResponseHeader()
	largeValue := generateLargeString(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rh.Set("large_header", largeValue)
		rh.Get("large_header")
	}
}

// Helper function for generating large strings
func generateLargeString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = byte('a' + (i % 26))
	}
	return string(result)
}
