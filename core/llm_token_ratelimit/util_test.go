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
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/google/uuid"
)

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		validate func(string) bool
	}{
		{
			name:  "empty parts",
			parts: []string{},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "single part",
			parts: []string{"test"},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "multiple parts",
			parts: []string{"part1", "part2", "part3"},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "empty string parts",
			parts: []string{"", "", ""},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "mixed empty and non-empty",
			parts: []string{"", "test", "", "data"},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "special characters",
			parts: []string{"test!@#$%^&*()", "unicode测试", "newline\nand\ttab"},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "long strings",
			parts: []string{strings.Repeat("a", 1000), strings.Repeat("b", 2000)},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
		{
			name:  "numeric strings",
			parts: []string{"123", "456.789", "-999"},
			validate: func(hash string) bool {
				return len(hash) > 0 && isHexString(hash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := generateHash(tt.parts...)
			if !tt.validate(hash) {
				t.Errorf("generateHash() = %v, validation failed", hash)
			}
		})
	}

	// Test consistency - same input should produce same hash
	t.Run("consistency", func(t *testing.T) {
		parts := []string{"test", "consistency", "check"}
		hash1 := generateHash(parts...)
		hash2 := generateHash(parts...)
		if hash1 != hash2 {
			t.Errorf("generateHash() inconsistent: %v != %v", hash1, hash2)
		}
	})

	// Test uniqueness - different inputs should produce different hashes
	t.Run("uniqueness", func(t *testing.T) {
		hash1 := generateHash("test1")
		hash2 := generateHash("test2")
		if hash1 == hash2 {
			t.Errorf("generateHash() not unique: %v == %v", hash1, hash2)
		}
	})

	// Test order sensitivity
	t.Run("order sensitivity", func(t *testing.T) {
		hash1 := generateHash("a", "b")
		hash2 := generateHash("b", "a")
		if hash1 == hash2 {
			t.Errorf("generateHash() order insensitive: %v == %v", hash1, hash2)
		}
	})
}

func TestParseRedisResponse(t *testing.T) {
	ctx := NewContext()
	ctx.Set(KeyRequestID, "test-request-123")

	tests := []struct {
		name     string
		response interface{}
		expected []int64
	}{
		{
			name:     "nil response",
			response: nil,
			expected: nil,
		},
		{
			name:     "non-slice response",
			response: "not a slice",
			expected: nil,
		},
		{
			name:     "empty slice",
			response: []interface{}{},
			expected: []int64{},
		},
		{
			name:     "int64 values",
			response: []interface{}{int64(1), int64(2), int64(3)},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "string values",
			response: []interface{}{"123", "456", "789"},
			expected: []int64{123, 456, 789},
		},
		{
			name:     "int values",
			response: []interface{}{1, 2, 3},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "float64 values",
			response: []interface{}{1.0, 2.0, 3.0},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "mixed valid types",
			response: []interface{}{int64(1), "2", 3, 4.0},
			expected: []int64{1, 2, 3, 4},
		},
		{
			name:     "negative numbers",
			response: []interface{}{-1, "-2", int64(-3), -4.0},
			expected: []int64{-1, -2, -3, -4},
		},
		{
			name:     "zero values",
			response: []interface{}{0, "0", int64(0), 0.0},
			expected: []int64{0, 0, 0, 0},
		},
		{
			name:     "large numbers",
			response: []interface{}{"9223372036854775807", int64(9223372036854775807)},
			expected: []int64{9223372036854775807, 9223372036854775807},
		},
		{
			name:     "invalid string number",
			response: []interface{}{"not_a_number"},
			expected: nil,
		},
		{
			name:     "invalid type",
			response: []interface{}{[]string{"nested", "array"}},
			expected: nil,
		},
		{
			name:     "mixed valid and invalid",
			response: []interface{}{1, "invalid", 3},
			expected: nil,
		},
		{
			name:     "float with decimals",
			response: []interface{}{1.7, 2.9, 3.1},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "string with leading/trailing spaces",
			response: []interface{}{" 123 ", "456"},
			expected: nil, // strconv.ParseInt doesn't handle spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisResponse(ctx, tt.response)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseRedisResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseRedisResponse_NilContext(t *testing.T) {
	response := []interface{}{1, 2, 3}
	result := parseRedisResponse(nil, response)
	expected := []int64{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("parseRedisResponse() with nil context = %v, want %v", result, expected)
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		validate func(string, int) bool
	}{
		{
			name: "zero length",
			n:    0,
			validate: func(s string, n int) bool {
				return s == ""
			},
		},
		{
			name: "negative length",
			n:    -5,
			validate: func(s string, n int) bool {
				return s == ""
			},
		},
		{
			name: "length 1",
			n:    1,
			validate: func(s string, n int) bool {
				return len(s) == 1 && isValidRandomChar(s[0])
			},
		},
		{
			name: "length 10",
			n:    10,
			validate: func(s string, n int) bool {
				return len(s) == 10 && allValidRandomChars(s)
			},
		},
		{
			name: "length 100",
			n:    100,
			validate: func(s string, n int) bool {
				return len(s) == 100 && allValidRandomChars(s)
			},
		},
		{
			name: "length 1000",
			n:    1000,
			validate: func(s string, n int) bool {
				return len(s) == 1000 && allValidRandomChars(s)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRandomString(tt.n)
			if !tt.validate(result, tt.n) {
				t.Errorf("generateRandomString(%d) = %v, validation failed", tt.n, result)
			}
		})
	}

	// Test uniqueness
	t.Run("uniqueness", func(t *testing.T) {
		const iterations = 1000
		const length = 20
		seen := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			s := generateRandomString(length)
			if seen[s] {
				t.Errorf("generateRandomString() produced duplicate: %s", s)
			}
			seen[s] = true
		}
	})

	// Test character distribution
	t.Run("character distribution", func(t *testing.T) {
		const length = 10000
		s := generateRandomString(length)
		charCount := make(map[byte]int)

		for i := 0; i < len(s); i++ {
			charCount[s[i]]++
		}

		// Should have reasonable distribution (not perfect, but not too skewed)
		if len(charCount) < 10 { // Should use more than 10 different characters
			t.Errorf("generateRandomString() poor character distribution: only %d unique chars", len(charCount))
		}
	})
}

func TestGenerateUUID(t *testing.T) {
	tests := []struct {
		name     string
		validate func(string) bool
	}{
		{
			name: "valid uuid format",
			validate: func(s string) bool {
				_, err := uuid.Parse(s)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUUID()
			if !tt.validate(result) {
				t.Errorf("generateUUID() = %v, validation failed", result)
			}
		})
	}

	// Test uniqueness
	t.Run("uniqueness", func(t *testing.T) {
		const iterations = 1000
		seen := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			uuid := generateUUID()
			if seen[uuid] {
				t.Errorf("generateUUID() produced duplicate: %s", uuid)
			}
			seen[uuid] = true
		}
	})

	// Test format consistency
	t.Run("format consistency", func(t *testing.T) {
		const iterations = 100

		for i := 0; i < iterations; i++ {
			uuid := generateUUID()
			if len(uuid) != 36 { // Standard UUID length
				t.Errorf("generateUUID() invalid length: %d, expected 36", len(uuid))
			}

			// Check hyphen positions
			if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
				t.Errorf("generateUUID() invalid format: %s", uuid)
			}
		}
	})
}

func TestDeepCopyByCopier(t *testing.T) {
	tests := []struct {
		name    string
		src     interface{}
		dest    interface{}
		wantErr bool
		verify  func(interface{}, interface{}) bool
	}{
		{
			name: "simple struct",
			src: &Rule{
				ID:       "test-rule-1",
				Resource: "test-resource",
				Strategy: FixedWindow,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-3.5-turbo",
				},
			},
			dest:    &Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRule := src.(*Rule)
				destRule := dest.(*Rule)
				return srcRule.ID == destRule.ID &&
					srcRule.Resource == destRule.Resource &&
					srcRule.Strategy == destRule.Strategy &&
					srcRule.Encoding.Provider == destRule.Encoding.Provider &&
					srcRule.Encoding.Model == destRule.Encoding.Model
			},
		},
		{
			name: "rule slice - simple rules",
			src: []*Rule{
				{
					ID:       "rule-1",
					Resource: "resource-1",
					Strategy: FixedWindow,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-3.5-turbo",
					},
				},
				{
					ID:       "rule-2",
					Resource: "resource-2",
					Strategy: PETA,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-4",
					},
				},
			},
			dest:    &[]*Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRules := src.([]*Rule)
				destRules := *(dest.(*[]*Rule))

				if len(srcRules) != len(destRules) {
					return false
				}

				for i, srcRule := range srcRules {
					destRule := destRules[i]
					if srcRule.ID != destRule.ID ||
						srcRule.Resource != destRule.Resource ||
						srcRule.Strategy != destRule.Strategy ||
						srcRule.Encoding.Provider != destRule.Encoding.Provider ||
						srcRule.Encoding.Model != destRule.Encoding.Model {
						return false
					}
				}
				return true
			},
		},
		{
			name: "rule slice - complex rules with specific items",
			src: []*Rule{
				{
					ID:       "complex-rule-1",
					Resource: "llm-api",
					Strategy: FixedWindow,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-3.5-turbo",
					},
					SpecificItems: []*SpecificItem{
						{
							Identifier: Identifier{
								Type:  Header,
								Value: "user-id",
							},
							KeyItems: []*KeyItem{
								{
									Key: "tokens-per-minute",
									Token: Token{
										Number:        1000,
										CountStrategy: TotalTokens,
									},
									Time: Time{
										Unit:  Minute,
										Value: 1,
									},
								},
								{
									Key: "tokens-per-hour",
									Token: Token{
										Number:        10000,
										CountStrategy: InputTokens,
									},
									Time: Time{
										Unit:  Hour,
										Value: 1,
									},
								},
							},
						},
						{
							Identifier: Identifier{
								Type:  AllIdentifier,
								Value: "*",
							},
							KeyItems: []*KeyItem{
								{
									Key: "global-limit",
									Token: Token{
										Number:        50000,
										CountStrategy: OutputTokens,
									},
									Time: Time{
										Unit:  Day,
										Value: 1,
									},
								},
							},
						},
					},
				},
				{
					ID:       "complex-rule-2",
					Resource: "chat-api",
					Strategy: PETA,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-4",
					},
					SpecificItems: []*SpecificItem{
						{
							Identifier: Identifier{
								Type:  Header,
								Value: "app-id",
							},
							KeyItems: []*KeyItem{
								{
									Key: "premium-limit",
									Token: Token{
										Number:        100000,
										CountStrategy: TotalTokens,
									},
									Time: Time{
										Unit:  Hour,
										Value: 24,
									},
								},
							},
						},
					},
				},
			},
			dest:    &[]*Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRules := src.([]*Rule)
				destRules := *(dest.(*[]*Rule))

				if len(srcRules) != len(destRules) {
					return false
				}

				for i, srcRule := range srcRules {
					destRule := destRules[i]

					// Verify basic fields
					if srcRule.ID != destRule.ID ||
						srcRule.Resource != destRule.Resource ||
						srcRule.Strategy != destRule.Strategy ||
						srcRule.Encoding.Provider != destRule.Encoding.Provider ||
						srcRule.Encoding.Model != destRule.Encoding.Model {
						return false
					}

					// Verify SpecificItems
					if len(srcRule.SpecificItems) != len(destRule.SpecificItems) {
						return false
					}

					for j, srcItem := range srcRule.SpecificItems {
						destItem := destRule.SpecificItems[j]

						// Verify Identifier
						if srcItem.Identifier.Type != destItem.Identifier.Type ||
							srcItem.Identifier.Value != destItem.Identifier.Value {
							return false
						}

						// Verify KeyItems
						if len(srcItem.KeyItems) != len(destItem.KeyItems) {
							return false
						}

						for k, srcKeyItem := range srcItem.KeyItems {
							destKeyItem := destItem.KeyItems[k]

							if srcKeyItem.Key != destKeyItem.Key ||
								srcKeyItem.Token.Number != destKeyItem.Token.Number ||
								srcKeyItem.Token.CountStrategy != destKeyItem.Token.CountStrategy ||
								srcKeyItem.Time.Unit != destKeyItem.Time.Unit ||
								srcKeyItem.Time.Value != destKeyItem.Time.Value {
								return false
							}
						}
					}
				}
				return true
			},
		},
		{
			name:    "empty rule slice",
			src:     []*Rule{},
			dest:    &[]*Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRules := src.([]*Rule)
				destRules := *(dest.(*[]*Rule))
				return len(srcRules) == 0 && len(destRules) == 0
			},
		},
		{
			name:    "nil-rule-slice-1",
			src:     nil,
			dest:    &[]*Rule{},
			wantErr: true,
			verify:  nil,
		},
		{
			name:    "nil-rule-slice-2",
			src:     []*Rule{nil},
			dest:    &[]*Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRules := src.([]*Rule)
				destRules := *(dest.(*[]*Rule))
				return srcRules[0] == nil && destRules[0] == nil
			},
		},
		{
			name: "rule slice with nil elements",
			src: []*Rule{
				{
					ID:       "rule-1",
					Resource: "resource-1",
					Strategy: FixedWindow,
				},
				nil,
				{
					ID:       "rule-3",
					Resource: "resource-3",
					Strategy: PETA,
				},
			},
			dest:    &[]*Rule{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				srcRules := src.([]*Rule)
				destRules := *(dest.(*[]*Rule))

				if len(srcRules) != len(destRules) {
					return false
				}

				// Check that nil elements are preserved
				if destRules[1] != nil {
					return false
				}

				// Check non-nil elements
				if srcRules[0].ID != destRules[0].ID ||
					srcRules[2].ID != destRules[2].ID {
					return false
				}

				return true
			},
		},
		{
			name:    "map to map",
			src:     map[string]interface{}{"key1": "value1", "key2": 42},
			dest:    &map[string]interface{}{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				return reflect.DeepEqual(src, *(dest.(*map[string]interface{})))
			},
		},
		{
			name:    "slice to slice",
			src:     []int{1, 2, 3, 4, 5},
			dest:    &[]int{},
			wantErr: false,
			verify: func(src, dest interface{}) bool {
				return reflect.DeepEqual(src, *(dest.(*[]int)))
			},
		},
		{
			name:    "invalid dest type",
			src:     "test",
			dest:    "not_a_pointer",
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := deepCopyByCopier(tt.src, tt.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("deepCopyByCopier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil && !tt.verify(tt.src, tt.dest) {
				t.Errorf("deepCopyByCopier() verification failed")
			}
		})
	}
}

func TestDeepCopyByCopier_RuleSliceMemoryIsolation(t *testing.T) {
	// Test that the copied slice is independent from the original
	original := []*Rule{
		{
			ID:       "rule-1",
			Resource: "resource-1",
			Strategy: FixedWindow,
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-3.5-turbo",
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  Header,
						Value: "user-id",
					},
					KeyItems: []*KeyItem{
						{
							Key: "test-key",
							Token: Token{
								Number:        1000,
								CountStrategy: TotalTokens,
							},
							Time: Time{
								Unit:  Minute,
								Value: 1,
							},
						},
					},
				},
			},
		},
	}

	copied := &[]*Rule{}
	err := deepCopyByCopier(original, copied)
	if err != nil {
		t.Fatalf("deepCopyByCopier() error = %v", err)
	}

	// Modify the original
	original[0].ID = "modified-rule-1"
	original[0].Resource = "modified-resource"
	original[0].SpecificItems[0].Identifier.Value = "modified-user-id"
	original[0].SpecificItems[0].KeyItems[0].Key = "modified-key"

	// Check that the copy is not affected
	copiedRules := *copied
	if copiedRules[0].ID != "rule-1" {
		t.Errorf("Copy was affected by original modification: ID = %v", copiedRules[0].ID)
	}
	if copiedRules[0].Resource != "resource-1" {
		t.Errorf("Copy was affected by original modification: Resource = %v", copiedRules[0].Resource)
	}
	if copiedRules[0].SpecificItems[0].Identifier.Value != "user-id" {
		t.Errorf("Copy was affected by original modification: Identifier.Value = %v",
			copiedRules[0].SpecificItems[0].Identifier.Value)
	}
	if copiedRules[0].SpecificItems[0].KeyItems[0].Key != "test-key" {
		t.Errorf("Copy was affected by original modification: KeyItem.Key = %v",
			copiedRules[0].SpecificItems[0].KeyItems[0].Key)
	}
}

func TestDeepCopyByCopier_LargeRuleSlice(t *testing.T) {
	// Test with a large number of rules
	const numRules = 1000
	original := make([]*Rule, numRules)

	for i := 0; i < numRules; i++ {
		original[i] = &Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Resource: fmt.Sprintf("resource-%d", i),
			Strategy: Strategy(i % 2), // Alternate between FixedWindow and PETA
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    fmt.Sprintf("model-%d", i),
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  IdentifierType(i % 2), // Alternate between AllIdentifier and Header
						Value: fmt.Sprintf("identifier-%d", i),
					},
					KeyItems: []*KeyItem{
						{
							Key: fmt.Sprintf("key-%d", i),
							Token: Token{
								Number:        int64(i * 100),
								CountStrategy: CountStrategy(i % 3), // Cycle through count strategies
							},
							Time: Time{
								Unit:  TimeUnit(i % 4), // Cycle through time units
								Value: int64(i + 1),
							},
						},
					},
				},
			},
		}
	}

	copied := &[]*Rule{}
	err := deepCopyByCopier(original, copied)
	if err != nil {
		t.Fatalf("deepCopyByCopier() error = %v", err)
	}

	copiedRules := *copied
	if len(copiedRules) != numRules {
		t.Fatalf("Expected %d rules, got %d", numRules, len(copiedRules))
	}

	// Verify all rules are correctly copied
	for i := 0; i < numRules; i++ {
		if copiedRules[i].ID != fmt.Sprintf("rule-%d", i) {
			t.Errorf("Rule %d ID mismatch: got %v", i, copiedRules[i].ID)
		}
		if copiedRules[i].Resource != fmt.Sprintf("resource-%d", i) {
			t.Errorf("Rule %d Resource mismatch: got %v", i, copiedRules[i].Resource)
		}
	}
}

// Concurrency tests
func TestGenerateHash_Concurrency(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([][]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			results[id] = make([]string, numOperations)

			for j := 0; j < numOperations; j++ {
				hash := generateHash(fmt.Sprintf("test_%d_%d", id, j))
				results[id][j] = hash
			}
		}(i)
	}

	wg.Wait()

	// Verify all hashes are unique
	seen := make(map[string]bool)
	for _, goroutineResults := range results {
		for _, hash := range goroutineResults {
			if seen[hash] {
				t.Errorf("Duplicate hash found: %s", hash)
			}
			seen[hash] = true
		}
	}
}

func TestGenerateRandomString_Concurrency(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 200
	const stringLength = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([][]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			results[id] = make([]string, numOperations)

			for j := 0; j < numOperations; j++ {
				s := generateRandomString(stringLength)
				results[id][j] = s
			}
		}(i)
	}

	wg.Wait()

	// Verify all strings are unique and valid
	seen := make(map[string]bool)
	for _, goroutineResults := range results {
		for _, s := range goroutineResults {
			if len(s) != stringLength {
				t.Errorf("Invalid string length: %d, expected %d", len(s), stringLength)
			}
			if !allValidRandomChars(s) {
				t.Errorf("Invalid characters in string: %s", s)
			}
			if seen[s] {
				t.Errorf("Duplicate string found: %s", s)
			}
			seen[s] = true
		}
	}
}

func TestDeepCopyByCopier_Concurrency(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 10

	originalRules := []*Rule{
		{
			ID:       "concurrent-rule",
			Resource: "concurrent-resource",
			Strategy: FixedWindow,
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-3.5-turbo",
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				copied := &[]*Rule{}
				err := deepCopyByCopier(originalRules, copied)
				if err != nil {
					t.Errorf("Goroutine %d, operation %d: deepCopyByCopier() error = %v", id, j, err)
					return
				}

				copiedRules := *copied
				if len(copiedRules) != 1 {
					t.Errorf("Goroutine %d, operation %d: expected 1 rule, got %d", id, j, len(copiedRules))
					return
				}

				if copiedRules[0].ID != "concurrent-rule" {
					t.Errorf("Goroutine %d, operation %d: wrong ID: %v", id, j, copiedRules[0].ID)
				}
			}
		}(i)
	}

	wg.Wait()
}

// Stress tests
func TestGenerateRandomString_Stress(t *testing.T) {
	// Generate many strings and check for basic properties
	const numStrings = 10000
	const stringLength = 50

	strings := make([]string, numStrings)
	for i := 0; i < numStrings; i++ {
		strings[i] = generateRandomString(stringLength)
	}

	// Check all strings are valid length
	for i, s := range strings {
		if len(s) != stringLength {
			t.Errorf("String %d has invalid length: %d", i, len(s))
		}
		if !utf8.ValidString(s) {
			t.Errorf("String %d is not valid UTF-8", i)
		}
	}

	// Check for reasonable uniqueness (allowing some duplicates due to randomness)
	unique := make(map[string]bool)
	for _, s := range strings {
		unique[s] = true
	}

	uniqueRatio := float64(len(unique)) / float64(numStrings)
	if uniqueRatio < 0.99 { // Expect at least 99% uniqueness
		t.Errorf("Poor uniqueness ratio: %f", uniqueRatio)
	}
}

func TestParseRedisResponse_Stress(t *testing.T) {
	ctx := NewContext()

	// Test with very large response
	const size = 100000
	response := make([]interface{}, size)
	for i := 0; i < size; i++ {
		response[i] = int64(i)
	}

	result := parseRedisResponse(ctx, response)
	if len(result) != size {
		t.Errorf("Expected %d results, got %d", size, len(result))
	}

	for i, val := range result {
		if val != int64(i) {
			t.Errorf("Unexpected value at index %d: %d", i, val)
		}
	}
}

func TestDeepCopyByCopier_RuleSlice_Stress(t *testing.T) {
	// Create a very large and complex rule slice
	const numRules = 5000
	original := make([]*Rule, numRules)

	for i := 0; i < numRules; i++ {
		original[i] = &Rule{
			ID:       fmt.Sprintf("stress-rule-%d", i),
			Resource: fmt.Sprintf("stress-resource-%d", i),
			Strategy: Strategy(i % 2),
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    fmt.Sprintf("stress-model-%d", i),
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  IdentifierType(i % 2),
						Value: fmt.Sprintf("stress-identifier-%d", i),
					},
					KeyItems: []*KeyItem{
						{
							Key: fmt.Sprintf("stress-key-%d", i),
							Token: Token{
								Number:        int64(i * 100),
								CountStrategy: CountStrategy(i % 3),
							},
							Time: Time{
								Unit:  TimeUnit(i % 4),
								Value: int64(i + 1),
							},
						},
					},
				},
			},
		}
	}

	copied := &[]*Rule{}
	err := deepCopyByCopier(original, copied)
	if err != nil {
		t.Fatalf("deepCopyByCopier() error = %v", err)
	}

	copiedRules := *copied
	if len(copiedRules) != numRules {
		t.Fatalf("Expected %d rules, got %d", numRules, len(copiedRules))
	}

	// Spot check some rules
	checkIndices := []int{0, numRules / 4, numRules / 2, 3 * numRules / 4, numRules - 1}
	for _, i := range checkIndices {
		if copiedRules[i].ID != fmt.Sprintf("stress-rule-%d", i) {
			t.Errorf("Rule %d ID mismatch: got %v", i, copiedRules[i].ID)
		}
		if copiedRules[i].Resource != fmt.Sprintf("stress-resource-%d", i) {
			t.Errorf("Rule %d Resource mismatch: got %v", i, copiedRules[i].Resource)
		}
	}
}

// Test edge cases for unsafe pointer usage in generateRandomString
func TestGenerateRandomString_UnsafePointer(t *testing.T) {
	// Test that the unsafe pointer conversion doesn't cause issues
	for i := 1; i <= 100; i++ {
		s := generateRandomString(i)
		if len(s) != i {
			t.Errorf("Length mismatch for size %d: got %d", i, len(s))
		}

		// Verify string is properly null-terminated and readable
		for j, r := range s {
			if r == 0 {
				t.Errorf("Unexpected null byte at position %d in string of length %d", j, i)
			}
		}
	}
}

// Performance tests
func BenchmarkGenerateHash_Single(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateHash("test_string")
	}
}

func BenchmarkGenerateHash_Multiple(b *testing.B) {
	parts := []string{"part1", "part2", "part3", "part4"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateHash(parts...)
	}
}

func BenchmarkGenerateHash_Large(b *testing.B) {
	largePart := strings.Repeat("a", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateHash(largePart, "part2", largePart)
	}
}

func BenchmarkParseRedisResponse(b *testing.B) {
	ctx := NewContext()
	response := []interface{}{int64(1), "2", 3, 4.0, int64(5)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRedisResponse(ctx, response)
	}
}

func BenchmarkParseRedisResponse_Large(b *testing.B) {
	ctx := NewContext()
	response := make([]interface{}, 1000)
	for i := range response {
		response[i] = int64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRedisResponse(ctx, response)
	}
}

func BenchmarkGenerateRandomString_Small(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateRandomString(10)
	}
}

func BenchmarkGenerateRandomString_Medium(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateRandomString(100)
	}
}

func BenchmarkGenerateRandomString_Large(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateRandomString(1000)
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateUUID()
	}
}

func BenchmarkDeepCopyByCopier_SimpleRule(b *testing.B) {
	src := &Rule{
		ID:       "benchmark-rule",
		Resource: "benchmark-resource",
		Strategy: FixedWindow,
		Encoding: TokenEncoding{
			Provider: OpenAIEncoderProvider,
			Model:    "gpt-3.5-turbo",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &Rule{}
		deepCopyByCopier(src, dest)
	}
}

func BenchmarkDeepCopyByCopier_RuleSlice_Small(b *testing.B) {
	src := []*Rule{
		{
			ID:       "rule-1",
			Resource: "resource-1",
			Strategy: FixedWindow,
		},
		{
			ID:       "rule-2",
			Resource: "resource-2",
			Strategy: PETA,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &[]*Rule{}
		deepCopyByCopier(src, dest)
	}
}

func BenchmarkDeepCopyByCopier_RuleSlice_Large(b *testing.B) {
	src := make([]*Rule, 100)
	for i := 0; i < 100; i++ {
		src[i] = &Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Resource: fmt.Sprintf("resource-%d", i),
			Strategy: Strategy(i % 2),
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    fmt.Sprintf("model-%d", i),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &[]*Rule{}
		deepCopyByCopier(src, dest)
	}
}

func BenchmarkDeepCopyByCopier_ComplexRuleSlice(b *testing.B) {
	src := []*Rule{
		{
			ID:       "complex-rule",
			Resource: "complex-resource",
			Strategy: FixedWindow,
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-4",
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  Header,
						Value: "user-id",
					},
					KeyItems: []*KeyItem{
						{
							Key: "tokens-per-minute",
							Token: Token{
								Number:        1000,
								CountStrategy: TotalTokens,
							},
							Time: Time{
								Unit:  Minute,
								Value: 1,
							},
						},
						{
							Key: "tokens-per-hour",
							Token: Token{
								Number:        10000,
								CountStrategy: InputTokens,
							},
							Time: Time{
								Unit:  Hour,
								Value: 1,
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &[]*Rule{}
		deepCopyByCopier(src, dest)
	}
}

// Parallel benchmarks
func BenchmarkGenerateHash_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateHash("test_string", "part2")
		}
	})
}

func BenchmarkGenerateRandomString_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateRandomString(20)
		}
	})
}

func BenchmarkGenerateUUID_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateUUID()
		}
	})
}

func BenchmarkDeepCopyByCopier_RuleSlice_Parallel(b *testing.B) {
	src := []*Rule{
		{
			ID:       "parallel-rule",
			Resource: "parallel-resource",
			Strategy: FixedWindow,
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dest := &[]*Rule{}
			deepCopyByCopier(src, dest)
		}
	})
}

// Memory allocation benchmarks
func BenchmarkGenerateHash_Memory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateHash("test", "memory", "benchmark")
	}
}

func BenchmarkGenerateRandomString_Memory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateRandomString(50)
	}
}

func BenchmarkParseRedisResponse_Memory(b *testing.B) {
	ctx := NewContext()
	response := []interface{}{int64(1), "2", 3, 4.0}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRedisResponse(ctx, response)
	}
}

func BenchmarkDeepCopyByCopier_Memory(b *testing.B) {
	src := []*Rule{
		{
			ID:       "memory-rule",
			Resource: "memory-resource",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &[]*Rule{}
		deepCopyByCopier(src, dest)
	}
}

// Helper functions
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isValidRandomChar(b byte) bool {
	for i := 0; i < len(RandomLetterBytes); i++ {
		if b == RandomLetterBytes[i] {
			return true
		}
	}
	return false
}

func allValidRandomChars(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isValidRandomChar(s[i]) {
			return false
		}
	}
	return true
}
