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
	"sync"
	"testing"
)

func TestAllIdentifierChecker_Check(t *testing.T) {
	tests := []struct {
		name       string
		checker    *AllIdentifierChecker
		ctx        *Context
		infos      *RequestInfos
		identifier Identifier
		pattern    string
		want       bool
	}{
		{
			name:    "nil checker",
			checker: nil,
			ctx:     NewContext(),
			infos:   &RequestInfos{},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    false,
		},
		{
			name:    "nil infos - global rate limit",
			checker: &AllIdentifierChecker{},
			ctx:     NewContext(),
			infos:   nil,
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    true,
		},
		{
			name:    "header checker matches",
			checker: &AllIdentifierChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    true,
		},
		{
			name:    "no checker matches",
			checker: &AllIdentifierChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"different"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    false,
		},
		{
			name:    "empty headers",
			checker: &AllIdentifierChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker.Check(tt.ctx, tt.infos, tt.identifier, tt.pattern)
			if got != tt.want {
				t.Errorf("AllIdentifierChecker.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderChecker_Check(t *testing.T) {
	tests := []struct {
		name       string
		checker    *HeaderChecker
		ctx        *Context
		infos      *RequestInfos
		identifier Identifier
		pattern    string
		want       bool
	}{
		{
			name:    "nil checker",
			checker: nil,
			ctx:     NewContext(),
			infos:   &RequestInfos{},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    false,
		},
		{
			name:    "nil infos - global rate limit",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos:   nil,
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    true,
		},
		{
			name:    "nil headers - global rate limit",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: nil,
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    true,
		},
		{
			name:    "header key and value match",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    true,
		},
		{
			name:    "header key matches but value does not",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"different"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    false,
		},
		{
			name:    "empty header values uses pattern as value",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test",
			want:    true,
		},
		{
			name:    "multiple header values - first matches",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123", "different"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    true,
		},
		{
			name:    "multiple header values - second matches",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"different", "test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    true,
		},
		{
			name:    "multiple header values - none match",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"different1", "different2"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    false,
		},
		{
			name:    "wildcard header key matches",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"X-User-Id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: ".*User.*",
			},
			pattern: "test.*",
			want:    true,
		},
		{
			name:    "wildcard header key does not match",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"App-Id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: ".*User.*",
			},
			pattern: "test.*",
			want:    false,
		},
		{
			name:    "exact header key match",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test123",
			want:    true,
		},
		{
			name:    "case sensitive header key",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"user-id": {"test123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "test.*",
			want:    false,
		},
		{
			name:    "wildcard pattern matches all",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"anything"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: ".*",
			want:    true,
		},
		{
			name:    "multiple headers with different keys",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"user123"},
					"App-Id":  {"app456"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "App-Id",
			},
			pattern: "app.*",
			want:    true,
		},
		{
			name:    "numeric pattern matching",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {"12345"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "\\d+",
			want:    true,
		},
		{
			name:    "empty pattern matches empty value",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Id": {""},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Id",
			},
			pattern: "",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker.Check(tt.ctx, tt.infos, tt.identifier, tt.pattern)
			if got != tt.want {
				t.Errorf("HeaderChecker.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Edge cases tests
func TestHeaderChecker_Check_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		checker    *HeaderChecker
		ctx        *Context
		infos      *RequestInfos
		identifier Identifier
		pattern    string
		want       bool
	}{
		{
			name:    "special characters in header key",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"X-Custom-Header_123": {"value"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "X-Custom-Header_123",
			},
			pattern: "value",
			want:    true,
		},
		{
			name:    "special characters in pattern",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"Content-Type": {"application/json; charset=utf-8"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "Content-Type",
			},
			pattern: "application/json.*",
			want:    true,
		},
		{
			name:    "unicode characters",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"User-Name": {"用户123"},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "User-Name",
			},
			pattern: ".*123",
			want:    true,
		},
		{
			name:    "very long header value",
			checker: &HeaderChecker{},
			ctx:     NewContext(),
			infos: &RequestInfos{
				Headers: map[string][]string{
					"Long-Header": {generateLongString(1000)},
				},
			},
			identifier: Identifier{
				Type:  Header,
				Value: "Long-Header",
			},
			pattern: ".*",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker.Check(tt.ctx, tt.infos, tt.identifier, tt.pattern)
			if got != tt.want {
				t.Errorf("HeaderChecker.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Concurrency tests
func TestAllIdentifierChecker_Check_Concurrency(t *testing.T) {
	globalRuleMatcher = &RuleMatcher{
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{
			Header: &HeaderChecker{},
		},
	}

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "test.*"

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				result := checker.Check(ctx, infos, identifier, pattern)
				if !result {
					t.Errorf("Expected true but got false in concurrent test")
				}
			}
		}()
	}

	wg.Wait()
}

func TestHeaderChecker_Check_Concurrency(t *testing.T) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "test.*"

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				result := checker.Check(ctx, infos, identifier, pattern)
				if !result {
					t.Errorf("Expected true but got false in concurrent test")
				}
			}
		}()
	}

	wg.Wait()
}

// Performance tests
func BenchmarkAllIdentifierChecker_Check(b *testing.B) {
	globalRuleMatcher = &RuleMatcher{
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{
			Header: &HeaderChecker{},
		},
	}

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id":   {"user123"},
			"App-Id":    {"app456"},
			"X-Request": {"req789"},
			"X-Version": {"v1.0"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkHeaderChecker_Check(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id":   {"user123"},
			"App-Id":    {"app456"},
			"X-Request": {"req789"},
			"X-Version": {"v1.0"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkHeaderChecker_Check_MultipleValues(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"admin", "user", "guest", "operator", "user123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkHeaderChecker_Check_WildcardPattern(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"X-User-Id":    {"user123"},
			"X-App-Id":     {"app456"},
			"X-Request-Id": {"req789"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: ".*User.*",
	}
	pattern := "user.*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkHeaderChecker_Check_LargeHeaders(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()

	// Create a large number of headers
	headers := make(map[string][]string)
	for i := 0; i < 100; i++ {
		headers[fmt.Sprintf("Header-%d", i)] = []string{fmt.Sprintf("value-%d", i)}
	}

	infos := &RequestInfos{Headers: headers}
	identifier := Identifier{
		Type:  Header,
		Value: "Header-50",
	}
	pattern := "value-50"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkHeaderChecker_Check_ComplexPattern(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user-123-admin-456"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "^user-\\d+-[a-z]+-\\d+$"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkAllIdentifierChecker_Check_Parallel(b *testing.B) {
	globalRuleMatcher = &RuleMatcher{
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{
			Header: &HeaderChecker{},
		},
	}

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			checker.Check(ctx, infos, identifier, pattern)
		}
	})
}

func BenchmarkHeaderChecker_Check_Parallel(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			checker.Check(ctx, infos, identifier, pattern)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkHeaderChecker_Check_Memory(b *testing.B) {
	checker := &HeaderChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

func BenchmarkAllIdentifierChecker_Check_Memory(b *testing.B) {
	globalRuleMatcher = &RuleMatcher{
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{
			Header: &HeaderChecker{},
		},
	}

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "user.*"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx, infos, identifier, pattern)
	}
}

// Helper functions
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = byte('a' + (i % 26))
	}
	return string(result)
}

// Test with nil global rule matcher
func TestAllIdentifierChecker_Check_NilGlobalMatcher(t *testing.T) {
	originalMatcher := globalRuleMatcher
	defer func() {
		globalRuleMatcher = originalMatcher
	}()

	globalRuleMatcher = nil

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "test.*"

	// This should not panic and should return false
	result := checker.Check(ctx, infos, identifier, pattern)
	if !result {
		t.Error("Expected false when globalRuleMatcher is nil")
	}
}

// Test with empty identifier checkers
func TestAllIdentifierChecker_Check_EmptyCheckers(t *testing.T) {
	originalMatcher := globalRuleMatcher
	defer func() {
		globalRuleMatcher = originalMatcher
	}()

	globalRuleMatcher = &RuleMatcher{
		IdentifierCheckers: map[IdentifierType]IdentifierChecker{},
	}

	checker := &AllIdentifierChecker{}
	ctx := NewContext()
	infos := &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"test123"},
		},
	}
	identifier := Identifier{
		Type:  Header,
		Value: "User-Id",
	}
	pattern := "test.*"

	result := checker.Check(ctx, infos, identifier, pattern)
	if result {
		t.Error("Expected false when no identifier checkers are available")
	}
}
