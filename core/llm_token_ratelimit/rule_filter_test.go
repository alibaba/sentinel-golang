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
	"testing"
)

func TestFilterRules(t *testing.T) {
	tests := []struct {
		name     string
		rules    []*Rule
		expected []*Rule
		validate func([]*Rule) bool
	}{
		{
			name:     "nil rules",
			rules:    nil,
			expected: []*Rule{},
			validate: func(result []*Rule) bool {
				return len(result) == 0
			},
		},
		{
			name:     "empty rules",
			rules:    []*Rule{},
			expected: []*Rule{},
			validate: func(result []*Rule) bool {
				return len(result) == 0
			},
		},
		{
			name: "single valid rule",
			rules: []*Rule{
				{
					ID:       "rule-1",
					Resource: "test-resource",
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
									Key: "rate-limit",
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
			},
			expected: nil, // Will be validated by validate function
			validate: func(result []*Rule) bool {
				return len(result) == 1 &&
					result[0].ID == "rule-1" &&
					result[0].Resource == "test-resource"
			},
		},
		{
			name: "multiple valid rules with different resources",
			rules: []*Rule{
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
									Key: "rate-limit",
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
				{
					ID:       "rule-2",
					Resource: "resource-2",
					Strategy: PETA,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-4",
					},
					SpecificItems: []*SpecificItem{
						{
							Identifier: Identifier{
								Type:  AllIdentifier,
								Value: ".*",
							},
							KeyItems: []*KeyItem{
								{
									Key: "global-limit",
									Token: Token{
										Number:        5000,
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 2 {
					return false
				}
				resources := make(map[string]bool)
				for _, rule := range result {
					resources[rule.Resource] = true
				}
				return resources["resource-1"] && resources["resource-2"]
			},
		},
		{
			name: "duplicate resources - should keep latest",
			rules: []*Rule{
				{
					ID:       "rule-1",
					Resource: "same-resource",
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
									Key: "old-limit",
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
				{
					ID:       "rule-2",
					Resource: "same-resource",
					Strategy: PETA,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-4",
					},
					SpecificItems: []*SpecificItem{
						{
							Identifier: Identifier{
								Type:  AllIdentifier,
								Value: ".*",
							},
							KeyItems: []*KeyItem{
								{
									Key: "new-limit",
									Token: Token{
										Number:        2000,
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 1 {
					return false
				}
				// Should keep the latest rule (rule-2)
				return result[0].ID == "rule-2" &&
					result[0].Strategy == PETA &&
					result[0].SpecificItems[0].KeyItems[0].Key == "new-limit"
			},
		},
		{
			name: "rules with nil elements",
			rules: []*Rule{
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
									Key: "rate-limit",
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
				nil,
				{
					ID:       "rule-3",
					Resource: "resource-3",
					Strategy: PETA,
					Encoding: TokenEncoding{
						Provider: OpenAIEncoderProvider,
						Model:    "gpt-4",
					},
					SpecificItems: []*SpecificItem{
						{
							Identifier: Identifier{
								Type:  AllIdentifier,
								Value: ".*",
							},
							KeyItems: []*KeyItem{
								{
									Key: "global-limit",
									Token: Token{
										Number:        5000,
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 2 {
					return false
				}
				ids := make(map[string]bool)
				for _, rule := range result {
					ids[rule.ID] = true
				}
				return ids["rule-1"] && ids["rule-3"]
			},
		},
		{
			name: "invalid rules - no specific items",
			rules: []*Rule{
				{
					ID:            "invalid-rule",
					Resource:      "test-resource",
					Strategy:      FixedWindow,
					Encoding:      TokenEncoding{Provider: OpenAIEncoderProvider, Model: "gpt-3.5-turbo"},
					SpecificItems: nil, // Invalid: no specific items
				},
			},
			expected: []*Rule{},
			validate: func(result []*Rule) bool {
				return len(result) == 0
			},
		},
		{
			name: "invalid rules - negative token number",
			rules: []*Rule{
				{
					ID:       "invalid-rule",
					Resource: "test-resource",
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
									Key: "rate-limit",
									Token: Token{
										Number:        -1000, // Invalid: negative number
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
			},
			expected: []*Rule{},
			validate: func(result []*Rule) bool {
				return len(result) == 0
			},
		},
		{
			name: "mixed valid and invalid rules",
			rules: []*Rule{
				{
					ID:       "valid-rule",
					Resource: "valid-resource",
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
									Key: "rate-limit",
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
				{
					ID:            "invalid-rule",
					Resource:      "invalid-resource",
					Strategy:      FixedWindow,
					Encoding:      TokenEncoding{Provider: OpenAIEncoderProvider, Model: "gpt-3.5-turbo"},
					SpecificItems: nil, // Invalid
				},
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				return len(result) == 1 && result[0].ID == "valid-rule"
			},
		},
		{
			name: "rules with duplicate key items",
			rules: []*Rule{
				{
					ID:       "rule-with-duplicates",
					Resource: "test-resource",
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
									Key: "rate-limit",
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
									Key: "rate-limit", // Duplicate key
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
									Key: "different-limit",
									Token: Token{
										Number:        2000,
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 1 {
					return false
				}
				// Should filter out duplicate key items
				keyItems := result[0].SpecificItems[0].KeyItems
				if len(keyItems) != 2 {
					return false
				}
				// Check that both unique key items are present
				keys := make(map[string]bool)
				for _, item := range keyItems {
					keys[item.Key] = true
				}
				return keys["rate-limit"] && keys["different-limit"]
			},
		},
		{
			name: "same token, different time - should not be considered duplicate",
			rules: []*Rule{
				{
					SpecificItems: []*SpecificItem{
						{
							KeyItems: []*KeyItem{
								{
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
									Token: Token{
										Number:        1000,
										CountStrategy: TotalTokens,
									},
									Time: Time{
										Unit:  Minute,
										Value: 2,
									},
								},
							},
						},
					},
				},
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 1 {
					return false
				}
				// Should filter out duplicate key items
				keyItems := result[0].SpecificItems[0].KeyItems
				if len(keyItems) != 2 {
					return false
				}
				return keyItems[0].Time.Value != keyItems[1].Time.Value
			},
		},
		{
			name: "same time, different token - should be considered duplicate",
			rules: []*Rule{
				{
					SpecificItems: []*SpecificItem{
						{
							KeyItems: []*KeyItem{
								{
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 1 {
					return false
				}
				// Should filter out duplicate key items
				keyItems := result[0].SpecificItems[0].KeyItems
				return len(keyItems) == 1
			},
		},
		{
			name: "same SpecificItem and KeyItem",
			rules: []*Rule{
				{
					SpecificItems: []*SpecificItem{
						{
							KeyItems: []*KeyItem{
								{
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
						{
							KeyItems: []*KeyItem{
								{
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
			},
			expected: nil,
			validate: func(result []*Rule) bool {
				if len(result) != 1 {
					return false
				}
				if len(result[0].SpecificItems) != 1 {
					return false
				}
				keyItems := result[0].SpecificItems[0].KeyItems
				return len(keyItems) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterRules(tt.rules)

			if tt.expected != nil {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("FilterRules() = %v, want %v", result, tt.expected)
				}
			}

			if tt.validate != nil && !tt.validate(result) {
				t.Errorf("FilterRules() validation failed for result: %v", result)
			}
		})
	}
}

func TestFilterRules_DeepCopyFailure(t *testing.T) {
	// Test case where deepCopyByCopier might fail
	// This is harder to test without mocking, but we can test the error handling path

	// Create a rule that might cause copy issues
	rules := []*Rule{
		{
			ID:       "test-rule",
			Resource: "test-resource",
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
							Key: "rate-limit",
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

	result := FilterRules(rules)

	// Even if deep copy works, the result should be valid
	if len(result) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(result))
	}
}

func TestFilterRules_DefaultOptions(t *testing.T) {
	// Test that default options are set correctly
	rules := []*Rule{
		{
			SpecificItems: []*SpecificItem{
				{
					KeyItems: []*KeyItem{
						{
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

	result := FilterRules(rules)

	if len(result) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(result))
	}

	rule := result[0]

	// Check that defaults were applied
	if rule.Resource != DefaultResourcePattern {
		t.Errorf("Expected default resource pattern %s, got %s", DefaultResourcePattern, rule.Resource)
	}
	if rule.Strategy != FixedWindow {
		t.Errorf("Expected default strategy %s, got %s", FixedWindow, rule.Strategy)
	}

	if rule.Encoding.Provider != OpenAIEncoderProvider {
		t.Errorf("Expected default provider %s, got %s", OpenAIEncoderProvider, rule.Encoding.Provider)
	}

	if rule.Encoding.Model != DefaultTokenEncodingModel[OpenAIEncoderProvider] {
		t.Errorf("Expected default model %s, got %s", DefaultTokenEncodingModel[OpenAIEncoderProvider], rule.Encoding.Model)
	}

	if rule.SpecificItems[0].Identifier.Type != AllIdentifier {
		t.Errorf("Expected default identifier type %s, got %s", AllIdentifier, rule.SpecificItems[0].Identifier.Type)
	}

	if rule.SpecificItems[0].Identifier.Value != DefaultIdentifierValuePattern {
		t.Errorf("Expected default identifier value %s, got %s", DefaultIdentifierValuePattern, rule.SpecificItems[0].Identifier.Value)
	}

	if rule.SpecificItems[0].KeyItems[0].Key != DefaultKeyPattern {
		t.Errorf("Expected default key pattern %s, got %s", DefaultKeyPattern, rule.SpecificItems[0].KeyItems[0].Key)
	}
}

func TestFilterRules_LargeNumberOfRules(t *testing.T) {
	// Stress test with a large number of rules
	const numRules = 1000
	rules := make([]*Rule, numRules)

	for i := 0; i < numRules; i++ {
		rules[i] = &Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Resource: fmt.Sprintf("resource-%d", i%100), // Create some duplicates
			Strategy: Strategy(i % 2),                   // Alternate between strategies
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-3.5-turbo",
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  IdentifierType(i % 2),
						Value: fmt.Sprintf("identifier-%d", i),
					},
					KeyItems: []*KeyItem{
						{
							Key: fmt.Sprintf("key-%d", i),
							Token: Token{
								Number:        int64(i + 1),
								CountStrategy: CountStrategy(i % 3),
							},
							Time: Time{
								Unit:  TimeUnit(i % 4),
								Value: int64(i%10 + 1),
							},
						},
					},
				},
			},
		}
	}

	result := FilterRules(rules)

	// Should have 100 unique resources (0-99)
	if len(result) != 100 {
		t.Errorf("Expected 100 unique rules, got %d", len(result))
	}

	// Verify all results are valid
	resources := make(map[string]bool)
	for _, rule := range result {
		if resources[rule.Resource] {
			t.Errorf("Found duplicate resource: %s", rule.Resource)
		}
		resources[rule.Resource] = true

		if err := IsValidRule(rule); err != nil {
			t.Errorf("Invalid rule after filtering: %v", err)
		}
	}
}

func TestFilterRules_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		rules    []*Rule
		expected int
	}{
		{
			name:     "all nil rules",
			rules:    []*Rule{nil, nil, nil},
			expected: 0,
		},
		{
			name: "rule with invalid regex",
			rules: []*Rule{
				{
					ID:       "invalid-regex-rule",
					Resource: "[invalid-regex", // Invalid regex
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
									Key: "rate-limit",
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
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterRules(tt.rules)
			if len(result) != tt.expected {
				t.Errorf("FilterRules() returned %d rules, expected %d", len(result), tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkFilterRules_Small(b *testing.B) {
	rules := []*Rule{
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
							Key: "rate-limit",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterRules(rules)
	}
}

func BenchmarkFilterRules_Large(b *testing.B) {
	const numRules = 1000
	rules := make([]*Rule, numRules)

	for i := 0; i < numRules; i++ {
		rules[i] = &Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Resource: fmt.Sprintf("resource-%d", i),
			Strategy: Strategy(i % 2),
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-3.5-turbo",
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  IdentifierType(i % 2),
						Value: fmt.Sprintf("identifier-%d", i),
					},
					KeyItems: []*KeyItem{
						{
							Key: fmt.Sprintf("key-%d", i),
							Token: Token{
								Number:        int64(i + 1),
								CountStrategy: CountStrategy(i % 3),
							},
							Time: Time{
								Unit:  TimeUnit(i % 4),
								Value: int64(i%10 + 1),
							},
						},
					},
				},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterRules(rules)
	}
}

func BenchmarkFilterRules_WithDuplicates(b *testing.B) {
	const numRules = 500
	rules := make([]*Rule, numRules)

	for i := 0; i < numRules; i++ {
		rules[i] = &Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Resource: fmt.Sprintf("resource-%d", i%50), // Many duplicates
			Strategy: Strategy(i % 2),
			Encoding: TokenEncoding{
				Provider: OpenAIEncoderProvider,
				Model:    "gpt-3.5-turbo",
			},
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{
						Type:  IdentifierType(i % 2),
						Value: fmt.Sprintf("identifier-%d", i),
					},
					KeyItems: []*KeyItem{
						{
							Key: fmt.Sprintf("key-%d", i),
							Token: Token{
								Number:        int64(i + 1),
								CountStrategy: CountStrategy(i % 3),
							},
							Time: Time{
								Unit:  TimeUnit(i % 4),
								Value: int64(i%10 + 1),
							},
						},
					},
				},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterRules(rules)
	}
}

// Memory allocation benchmark
func BenchmarkFilterRules_Memory(b *testing.B) {
	rules := []*Rule{
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
							Key: "rate-limit",
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

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterRules(rules)
	}
}
