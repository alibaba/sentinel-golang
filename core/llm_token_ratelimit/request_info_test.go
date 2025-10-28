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
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestWithHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected map[string][]string
	}{
		{
			name: "normal headers",
			headers: map[string][]string{
				"User-Id": {"test123"},
				"App-Id":  {"app456"},
			},
			expected: map[string][]string{
				"User-Id": {"test123"},
				"App-Id":  {"app456"},
			},
		},
		{
			name:     "empty headers",
			headers:  map[string][]string{},
			expected: map[string][]string{},
		},
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
		{
			name: "single header with multiple values",
			headers: map[string][]string{
				"Accept": {"application/json", "text/html"},
			},
			expected: map[string][]string{
				"Accept": {"application/json", "text/html"},
			},
		},
		{
			name: "header with empty value",
			headers: map[string][]string{
				"Empty-Header": {""},
			},
			expected: map[string][]string{
				"Empty-Header": {""},
			},
		},
		{
			name: "header with nil slice",
			headers: map[string][]string{
				"Nil-Values": nil,
			},
			expected: map[string][]string{
				"Nil-Values": nil,
			},
		},
		{
			name: "complex headers",
			headers: map[string][]string{
				"Content-Type":  {"application/json; charset=utf-8"},
				"Authorization": {"Bearer token123"},
				"Cache-Control": {"no-cache", "no-store"},
				"X-Custom-Info": {"custom_value"},
			},
			expected: map[string][]string{
				"Content-Type":  {"application/json; charset=utf-8"},
				"Authorization": {"Bearer token123"},
				"Cache-Control": {"no-cache", "no-store"},
				"X-Custom-Info": {"custom_value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := &RequestInfos{}
			headerFunc := WithHeader(tt.headers)
			headerFunc(infos)

			if !reflect.DeepEqual(infos.Headers, tt.expected) {
				t.Errorf("WithHeader() failed, got %v, want %v", infos.Headers, tt.expected)
			}
		})
	}
}

func TestWithPrompts(t *testing.T) {
	tests := []struct {
		name     string
		prompts  []string
		expected []string
	}{
		{
			name:     "normal prompts",
			prompts:  []string{"prompt1", "prompt2", "prompt3"},
			expected: []string{"prompt1", "prompt2", "prompt3"},
		},
		{
			name:     "single prompt",
			prompts:  []string{"single_prompt"},
			expected: []string{"single_prompt"},
		},
		{
			name:     "empty prompts",
			prompts:  []string{},
			expected: []string{},
		},
		{
			name:     "nil prompts",
			prompts:  nil,
			expected: nil,
		},
		{
			name:     "prompts with empty strings",
			prompts:  []string{"", "non_empty", ""},
			expected: []string{"", "non_empty", ""},
		},
		{
			name: "long prompts",
			prompts: []string{
				"This is a very long prompt that contains multiple words and sentences to test the handling of lengthy input data.",
				"Another long prompt with different content to ensure proper storage and retrieval of extended text data.",
			},
			expected: []string{
				"This is a very long prompt that contains multiple words and sentences to test the handling of lengthy input data.",
				"Another long prompt with different content to ensure proper storage and retrieval of extended text data.",
			},
		},
		{
			name: "prompts with special characters",
			prompts: []string{
				"Prompt with special chars: !@#$%^&*()",
				"Unicode prompt: 你好世界",
				"Newline\nand\ttab characters",
			},
			expected: []string{
				"Prompt with special chars: !@#$%^&*()",
				"Unicode prompt: 你好世界",
				"Newline\nand\ttab characters",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := &RequestInfos{}
			promptFunc := WithPrompts(tt.prompts)
			promptFunc(infos)

			if !reflect.DeepEqual(infos.Prompts, tt.expected) {
				t.Errorf("WithPrompts() failed, got %v, want %v", infos.Prompts, tt.expected)
			}
		})
	}
}

func TestGenerateRequestInfos(t *testing.T) {
	tests := []struct {
		name     string
		funcs    []RequestInfo
		expected *RequestInfos
	}{
		{
			name:     "empty funcs",
			funcs:    []RequestInfo{},
			expected: &RequestInfos{},
		},
		{
			name:     "nil funcs",
			funcs:    nil,
			expected: &RequestInfos{},
		},
		{
			name: "with headers only",
			funcs: []RequestInfo{
				WithHeader(map[string][]string{
					"User-Id": {"test123"},
				}),
			},
			expected: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
			},
		},
		{
			name: "with prompts only",
			funcs: []RequestInfo{
				WithPrompts([]string{"prompt1", "prompt2"}),
			},
			expected: &RequestInfos{
				Prompts: []string{"prompt1", "prompt2"},
			},
		},
		{
			name: "with both headers and prompts",
			funcs: []RequestInfo{
				WithHeader(map[string][]string{
					"User-Id": {"test123"},
					"App-Id":  {"app456"},
				}),
				WithPrompts([]string{"prompt1", "prompt2"}),
			},
			expected: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
					"App-Id":  {"app456"},
				},
				Prompts: []string{"prompt1", "prompt2"},
			},
		},
		{
			name: "with nil function in slice",
			funcs: []RequestInfo{
				WithHeader(map[string][]string{
					"User-Id": {"test123"},
				}),
				nil,
				WithPrompts([]string{"prompt1"}),
			},
			expected: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
				Prompts: []string{"prompt1"},
			},
		},
		{
			name: "overwrite headers",
			funcs: []RequestInfo{
				WithHeader(map[string][]string{
					"User-Id": {"test123"},
				}),
				WithHeader(map[string][]string{
					"App-Id": {"app456"},
				}),
			},
			expected: &RequestInfos{
				Headers: map[string][]string{
					"App-Id": {"app456"},
				},
			},
		},
		{
			name: "overwrite prompts",
			funcs: []RequestInfo{
				WithPrompts([]string{"prompt1", "prompt2"}),
				WithPrompts([]string{"prompt3"}),
			},
			expected: &RequestInfos{
				Prompts: []string{"prompt3"},
			},
		},
		{
			name: "multiple operations",
			funcs: []RequestInfo{
				WithHeader(map[string][]string{
					"Initial-Header": {"initial"},
				}),
				WithPrompts([]string{"initial_prompt"}),
				WithHeader(map[string][]string{
					"Final-Header": {"final"},
				}),
				WithPrompts([]string{"final_prompt1", "final_prompt2"}),
			},
			expected: &RequestInfos{
				Headers: map[string][]string{
					"Final-Header": {"final"},
				},
				Prompts: []string{"final_prompt1", "final_prompt2"},
			},
		},
		{
			name:     "all nil functions",
			funcs:    []RequestInfo{nil, nil, nil},
			expected: &RequestInfos{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateRequestInfos(tt.funcs...)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GenerateRequestInfos() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractRequestInfos(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *Context
		expected *RequestInfos
	}{
		{
			name:     "nil context",
			ctx:      nil,
			expected: nil,
		},
		{
			name:     "context with no request infos",
			ctx:      NewContext(),
			expected: nil,
		},
		{
			name: "context with valid request infos",
			ctx: func() *Context {
				ctx := NewContext()
				reqInfos := &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"test123"},
					},
					Prompts: []string{"prompt1"},
				}
				ctx.Set(KeyRequestInfos, reqInfos)
				return ctx
			}(),
			expected: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
				Prompts: []string{"prompt1"},
			},
		},
		{
			name: "context with invalid type for request infos",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, "invalid_type")
				return ctx
			}(),
			expected: nil,
		},
		{
			name: "context with nil request infos",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, (*RequestInfos)(nil))
				return ctx
			}(),
			expected: nil,
		},
		{
			name: "context with empty request infos",
			ctx: func() *Context {
				ctx := NewContext()
				reqInfos := &RequestInfos{}
				ctx.Set(KeyRequestInfos, reqInfos)
				return ctx
			}(),
			expected: &RequestInfos{},
		},
		{
			name: "context with complex request infos",
			ctx: func() *Context {
				ctx := NewContext()
				reqInfos := &RequestInfos{
					Headers: map[string][]string{
						"Content-Type":  {"application/json"},
						"Authorization": {"Bearer token"},
						"User-Agent":    {"test-client/1.0"},
					},
					Prompts: []string{
						"Tell me about AI",
						"What is machine learning?",
						"Explain neural networks",
					},
				}
				ctx.Set(KeyRequestInfos, reqInfos)
				return ctx
			}(),
			expected: &RequestInfos{
				Headers: map[string][]string{
					"Content-Type":  {"application/json"},
					"Authorization": {"Bearer token"},
					"User-Agent":    {"test-client/1.0"},
				},
				Prompts: []string{
					"Tell me about AI",
					"What is machine learning?",
					"Explain neural networks",
				},
			},
		},
		{
			name: "context with different key",
			ctx: func() *Context {
				ctx := NewContext()
				reqInfos := &RequestInfos{
					Headers: map[string][]string{
						"Test": {"value"},
					},
				}
				ctx.Set("different_key", reqInfos)
				return ctx
			}(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRequestInfos(tt.ctx)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractRequestInfos() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Edge cases and error handling tests
func TestGenerateRequestInfos_EdgeCases(t *testing.T) {
	// Test with variadic function edge cases
	t.Run("no arguments", func(t *testing.T) {
		result := GenerateRequestInfos()
		if result == nil {
			t.Fatal("GenerateRequestInfos() should not return nil")
		}
		if result.Headers != nil || result.Prompts != nil {
			t.Fatal("GenerateRequestInfos() should return empty RequestInfos")
		}
	})

	t.Run("large number of functions", func(t *testing.T) {
		funcs := make([]RequestInfo, 1000)
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				funcs[i] = WithHeader(map[string][]string{
					fmt.Sprintf("Header-%d", i): {fmt.Sprintf("value-%d", i)},
				})
			} else {
				funcs[i] = WithPrompts([]string{fmt.Sprintf("prompt-%d", i)})
			}
		}

		result := GenerateRequestInfos(funcs...)
		if result == nil {
			t.Fatal("GenerateRequestInfos() should not return nil")
		}
		// Last header and prompt should be set
		if result.Headers == nil || result.Prompts == nil {
			t.Fatal("Headers and Prompts should be set")
		}
	})
}

// Concurrency tests
func TestWithHeader_Concurrency(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				headers := map[string][]string{
					fmt.Sprintf("Header-%d-%d", id, j): {fmt.Sprintf("value-%d-%d", id, j)},
				}
				infos := &RequestInfos{}
				headerFunc := WithHeader(headers)
				headerFunc(infos)

				if infos.Headers == nil {
					t.Errorf("Headers should not be nil")
				}
			}
		}(i)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Test passed
	case <-time.After(10 * time.Second):
		t.Fatal("Test timeout")
	}
}

func TestGenerateRequestInfos_Concurrency(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 200

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				funcs := []RequestInfo{
					WithHeader(map[string][]string{
						fmt.Sprintf("User-Id-%d", id): {fmt.Sprintf("user-%d-%d", id, j)},
					}),
					WithPrompts([]string{fmt.Sprintf("prompt-%d-%d", id, j)}),
				}

				result := GenerateRequestInfos(funcs...)
				if result == nil {
					t.Error("GenerateRequestInfos should not return nil")
					return
				}
				if result.Headers == nil || result.Prompts == nil {
					t.Error("Headers and Prompts should be set")
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// Stress tests
func TestGenerateRequestInfos_StressTest(t *testing.T) {
	const numFunctions = 10000

	funcs := make([]RequestInfo, numFunctions)
	for i := 0; i < numFunctions; i++ {
		if i%2 == 0 {
			funcs[i] = WithHeader(map[string][]string{
				fmt.Sprintf("Header-%d", i): {fmt.Sprintf("value-%d", i)},
			})
		} else {
			funcs[i] = WithPrompts([]string{fmt.Sprintf("prompt-%d", i)})
		}
	}

	result := GenerateRequestInfos(funcs...)
	if result == nil {
		t.Fatal("GenerateRequestInfos should not return nil")
	}

	// Verify that the last values are set correctly
	expectedHeaderKey := fmt.Sprintf("Header-%d", numFunctions-2) // Last even index
	if result.Headers == nil || result.Headers[expectedHeaderKey] == nil {
		t.Fatal("Last header should be set")
	}

	expectedPrompt := fmt.Sprintf("prompt-%d", numFunctions-1) // Last odd index
	if len(result.Prompts) == 0 || result.Prompts[0] != expectedPrompt {
		t.Fatal("Last prompt should be set")
	}
}

// Performance tests
func BenchmarkWithHeader(b *testing.B) {
	headers := map[string][]string{
		"User-Id":   {"user123"},
		"App-Id":    {"app456"},
		"X-Request": {"req789"},
		"X-Version": {"v1.0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		infos := &RequestInfos{}
		headerFunc := WithHeader(headers)
		headerFunc(infos)
	}
}

func BenchmarkWithPrompts(b *testing.B) {
	prompts := []string{
		"Tell me about artificial intelligence",
		"What is machine learning?",
		"Explain deep learning concepts",
		"How do neural networks work?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		infos := &RequestInfos{}
		promptFunc := WithPrompts(prompts)
		promptFunc(infos)
	}
}

func BenchmarkGenerateRequestInfos(b *testing.B) {
	headers := map[string][]string{
		"User-Id": {"user123"},
		"App-Id":  {"app456"},
	}
	prompts := []string{"prompt1", "prompt2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRequestInfos(
			WithHeader(headers),
			WithPrompts(prompts),
		)
	}
}

func BenchmarkExtractRequestInfos(b *testing.B) {
	ctx := NewContext()
	reqInfos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
		Prompts: []string{"prompt1"},
	}
	ctx.Set(KeyRequestInfos, reqInfos)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractRequestInfos(ctx)
	}
}

func BenchmarkGenerateRequestInfos_Complex(b *testing.B) {
	headers := map[string][]string{
		"Content-Type":  {"application/json; charset=utf-8"},
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		"User-Agent":    {"test-client/1.0"},
		"Accept":        {"application/json", "text/plain"},
		"Cache-Control": {"no-cache"},
		"X-Request-ID":  {"req-12345"},
		"X-User-ID":     {"user-67890"},
		"X-Session-ID":  {"sess-abcdef"},
	}
	prompts := []string{
		"Explain the fundamentals of quantum computing and its potential applications in cryptography",
		"Describe the differences between supervised and unsupervised machine learning algorithms",
		"What are the key components of a neural network and how do they work together?",
		"Discuss the ethical implications of artificial intelligence in autonomous vehicles",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRequestInfos(
			WithHeader(headers),
			WithPrompts(prompts),
		)
	}
}

// Parallel benchmarks
func BenchmarkWithHeader_Parallel(b *testing.B) {
	headers := map[string][]string{
		"User-Id": {"user123"},
		"App-Id":  {"app456"},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			infos := &RequestInfos{}
			headerFunc := WithHeader(headers)
			headerFunc(infos)
		}
	})
}

func BenchmarkGenerateRequestInfos_Parallel(b *testing.B) {
	headers := map[string][]string{
		"User-Id": {"user123"},
		"App-Id":  {"app456"},
	}
	prompts := []string{"prompt1", "prompt2"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateRequestInfos(
				WithHeader(headers),
				WithPrompts(prompts),
			)
		}
	})
}

func BenchmarkExtractRequestInfos_Parallel(b *testing.B) {
	ctx := NewContext()
	reqInfos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
		Prompts: []string{"prompt1"},
	}
	ctx.Set(KeyRequestInfos, reqInfos)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			extractRequestInfos(ctx)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkWithHeader_Memory(b *testing.B) {
	headers := map[string][]string{
		"User-Id": {"user123"},
		"App-Id":  {"app456"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		infos := &RequestInfos{}
		headerFunc := WithHeader(headers)
		headerFunc(infos)
	}
}

func BenchmarkGenerateRequestInfos_Memory(b *testing.B) {
	headers := map[string][]string{
		"User-Id": {"user123"},
		"App-Id":  {"app456"},
	}
	prompts := []string{"prompt1", "prompt2"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRequestInfos(
			WithHeader(headers),
			WithPrompts(prompts),
		)
	}
}

func BenchmarkExtractRequestInfos_Memory(b *testing.B) {
	ctx := NewContext()
	reqInfos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
		Prompts: []string{"prompt1"},
	}
	ctx.Set(KeyRequestInfos, reqInfos)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractRequestInfos(ctx)
	}
}
