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

	"github.com/alibaba/sentinel-golang/core/base"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}
	if ctx.userContext == nil {
		t.Error("userContext should be initialized")
	}
	if len(ctx.userContext) != 0 {
		t.Error("userContext should be empty initially")
	}
}

func TestContext_Set(t *testing.T) {
	tests := []struct {
		name  string
		ctx   *Context
		key   string
		value interface{}
	}{
		{
			name:  "nil context",
			ctx:   nil,
			key:   "test",
			value: "value",
		},
		{
			name:  "normal set",
			ctx:   NewContext(),
			key:   "test",
			value: "value",
		},
		{
			name:  "set with nil userContext",
			ctx:   &Context{userContext: nil},
			key:   "test",
			value: "value",
		},
		{
			name:  "overwrite existing key",
			ctx:   NewContext(),
			key:   "test",
			value: "new_value",
		},
		{
			name:  "set nil value",
			ctx:   NewContext(),
			key:   "test",
			value: nil,
		},
		{
			name:  "set empty string key",
			ctx:   NewContext(),
			key:   "",
			value: "value",
		},
		{
			name:  "set complex value",
			ctx:   NewContext(),
			key:   "complex",
			value: map[string]interface{}{"nested": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "overwrite existing key" {
				tt.ctx.Set("test", "old_value")
			}

			// This should not panic
			tt.ctx.Set(tt.key, tt.value)

			if tt.ctx != nil {
				got := tt.ctx.Get(tt.key)
				if !reflect.DeepEqual(got, tt.value) {
					t.Errorf("Context.Set() failed, got %v, want %v", got, tt.value)
				}
			}
		})
	}
}

func TestContext_Get(t *testing.T) {
	tests := []struct {
		name string
		ctx  *Context
		key  string
		want interface{}
	}{
		{
			name: "nil context",
			ctx:  nil,
			key:  "test",
			want: nil,
		},
		{
			name: "nil userContext",
			ctx:  &Context{userContext: nil},
			key:  "test",
			want: nil,
		},
		{
			name: "existing key",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set("test", "value")
				return ctx
			}(),
			key:  "test",
			want: "value",
		},
		{
			name: "non-existing key",
			ctx:  NewContext(),
			key:  "nonexistent",
			want: nil,
		},
		{
			name: "empty string key",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set("", "empty_key_value")
				return ctx
			}(),
			key:  "",
			want: "empty_key_value",
		},
		{
			name: "nil value stored",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set("nil_value", nil)
				return ctx
			}(),
			key:  "nil_value",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.Get(tt.key)
			if got != tt.want {
				t.Errorf("Context.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_extractArgs(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *Context
		entryCtx *base.EntryContext
		wantSet  bool
	}{
		{
			name:     "nil context",
			ctx:      nil,
			entryCtx: &base.EntryContext{},
			wantSet:  false,
		},
		{
			name:     "nil entryCtx",
			ctx:      NewContext(),
			entryCtx: nil,
			wantSet:  false,
		},
		{
			name: "nil input",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: nil,
			},
			wantSet: false,
		},
		{
			name: "nil args",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: nil,
				},
			},
			wantSet: false,
		},
		{
			name: "empty args",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{},
				},
			},
			wantSet: false,
		},
		{
			name: "valid RequestInfos in args",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						&RequestInfos{
							Headers: map[string][]string{
								"User-Id": {"test"},
							},
						},
					},
				},
			},
			wantSet: true,
		},
		{
			name: "invalid args type",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						"not_request_infos",
					},
				},
			},
			wantSet: false,
		},
		{
			name: "nil RequestInfos in args",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						(*RequestInfos)(nil),
					},
				},
			},
			wantSet: false,
		},
		{
			name: "multiple args with RequestInfos",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						"string_arg",
						123,
						&RequestInfos{
							Headers: map[string][]string{
								"App-Id": {"app123"},
							},
						},
						"another_arg",
					},
				},
			},
			wantSet: true,
		},
		{
			name: "multiple RequestInfos in args",
			ctx:  NewContext(),
			entryCtx: &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						&RequestInfos{
							Headers: map[string][]string{
								"User-Id": {"user123"},
							},
						},
						&RequestInfos{
							Headers: map[string][]string{
								"App-Id": {"app456"},
							},
						},
					},
				},
			},
			wantSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx != nil {
				// Clear any existing data
				tt.ctx.userContext = make(map[string]interface{})
			}

			tt.ctx.extractArgs(tt.entryCtx)

			if tt.ctx != nil {
				got := tt.ctx.Get(KeyRequestInfos)
				if tt.wantSet && got == nil {
					t.Error("extractArgs() should have set RequestInfos but didn't")
				}
				if !tt.wantSet && got != nil {
					t.Error("extractArgs() should not have set RequestInfos but did")
				}

				// Verify the RequestInfos is properly set
				if tt.wantSet {
					if reqInfos, ok := got.(*RequestInfos); ok {
						if reqInfos == nil {
							t.Error("RequestInfos should not be nil")
						}
					} else {
						t.Error("stored value should be *RequestInfos type")
					}
				}
			}
		})
	}
}

// Concurrency tests
func TestContext_ConcurrentAccess(t *testing.T) {
	ctx := NewContext()
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // readers and writers

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				ctx.Set(key, value)
			}
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				ctx.Get(key)
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
		t.Fatal("Test timeout - possible deadlock")
	}
}

func TestContext_ConcurrentSetGet(t *testing.T) {
	ctx := NewContext()
	const numGoroutines = 50
	const numOperations = 200

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("shared_key_%d", id%10) // Use shared keys to increase contention

			for j := 0; j < numOperations; j++ {
				value := fmt.Sprintf("value_%d_%d", id, j)
				ctx.Set(key, value)

				retrieved := ctx.Get(key)
				if retrieved == nil {
					t.Errorf("Expected value but got nil for key %s", key)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestContext_ExtractArgsConcurrency(t *testing.T) {
	ctx := NewContext()
	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			entryCtx := &base.EntryContext{
				Input: &base.SentinelInput{
					Args: []interface{}{
						&RequestInfos{
							Headers: map[string][]string{
								"User-Id": {fmt.Sprintf("user_%d", id)},
							},
						},
					},
				},
			}

			ctx.extractArgs(entryCtx)
		}(i)
	}

	wg.Wait()

	// Verify that some RequestInfos was set
	reqInfos := ctx.Get(KeyRequestInfos)
	if reqInfos == nil {
		t.Error("Expected RequestInfos to be set after concurrent extractArgs calls")
	}
}

// Edge cases and stress tests
func TestContext_StressTest(t *testing.T) {
	ctx := NewContext()
	const numKeys = 10000
	const numGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()

			// Each goroutine works with its own set of keys
			start := goroutineID * (numKeys / numGoroutines)
			end := (goroutineID + 1) * (numKeys / numGoroutines)

			for i := start; i < end; i++ {
				key := fmt.Sprintf("stress_key_%d", i)
				value := fmt.Sprintf("stress_value_%d", i)

				ctx.Set(key, value)

				retrieved := ctx.Get(key)
				if retrieved != value {
					t.Errorf("Stress test failed: expected %s, got %v", value, retrieved)
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestContext_LargeDataHandling(t *testing.T) {
	ctx := NewContext()

	// Test with large data structures
	largeMap := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	ctx.Set("large_data", largeMap)

	retrieved := ctx.Get("large_data")
	if retrieved == nil {
		t.Error("Large data should be retrievable")
	}

	if retrievedMap, ok := retrieved.(map[string]interface{}); ok {
		if len(retrievedMap) != 1000 {
			t.Errorf("Expected 1000 items in large map, got %d", len(retrievedMap))
		}
	} else {
		t.Error("Retrieved data should be a map")
	}
}

// Test extractArgs with various argument types
func TestContext_extractArgs_VariousTypes(t *testing.T) {
	ctx := NewContext()

	// Test with mixed argument types
	entryCtx := &base.EntryContext{
		Input: &base.SentinelInput{
			Args: []interface{}{
				"string_arg",
				123,
				3.14,
				true,
				[]string{"slice", "arg"},
				map[string]string{"map": "arg"},
				&RequestInfos{
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
				},
				struct{ Field string }{Field: "struct_arg"},
			},
		},
	}

	ctx.extractArgs(entryCtx)

	reqInfos := ctx.Get(KeyRequestInfos)
	if reqInfos == nil {
		t.Error("RequestInfos should be extracted from mixed arguments")
	}

	if ri, ok := reqInfos.(*RequestInfos); ok {
		if ri.Headers["Content-Type"][0] != "application/json" {
			t.Error("RequestInfos headers not properly extracted")
		}
	} else {
		t.Error("Extracted value should be *RequestInfos")
	}
}

// Performance tests
func BenchmarkContext_Set(b *testing.B) {
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Set("test_key", "test_value")
	}
}

func BenchmarkContext_Get(b *testing.B) {
	ctx := NewContext()
	ctx.Set("test_key", "test_value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Get("test_key")
	}
}

func BenchmarkContext_SetParallel(b *testing.B) {
	ctx := NewContext()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Set("test_key", "test_value")
		}
	})
}

func BenchmarkContext_GetParallel(b *testing.B) {
	ctx := NewContext()
	ctx.Set("test_key", "test_value")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Get("test_key")
		}
	})
}

func BenchmarkContext_SetGet(b *testing.B) {
	ctx := NewContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100) // Reuse keys to simulate real usage
		ctx.Set(key, i)
		ctx.Get(key)
	}
}

func BenchmarkContext_SetGetParallel(b *testing.B) {
	ctx := NewContext()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%100)
			ctx.Set(key, i)
			ctx.Get(key)
			i++
		}
	})
}

func BenchmarkContext_extractArgs(b *testing.B) {
	ctx := NewContext()
	entryCtx := &base.EntryContext{
		Input: &base.SentinelInput{
			Args: []interface{}{
				&RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"test"},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.extractArgs(entryCtx)
	}
}

func BenchmarkContext_extractArgsParallel(b *testing.B) {
	ctx := NewContext()
	entryCtx := &base.EntryContext{
		Input: &base.SentinelInput{
			Args: []interface{}{
				&RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"test"},
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.extractArgs(entryCtx)
		}
	})
}

func BenchmarkNewContext(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewContext()
	}
}

// Memory allocation benchmarks
func BenchmarkContext_Set_Memory(b *testing.B) {
	ctx := NewContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Set("test_key", "test_value")
	}
}

func BenchmarkContext_Get_Memory(b *testing.B) {
	ctx := NewContext()
	ctx.Set("test_key", "test_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Get("test_key")
	}
}

func BenchmarkContext_extractArgs_Memory(b *testing.B) {
	ctx := NewContext()
	entryCtx := &base.EntryContext{
		Input: &base.SentinelInput{
			Args: []interface{}{
				&RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"test"},
					},
				},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.extractArgs(entryCtx)
	}
}
