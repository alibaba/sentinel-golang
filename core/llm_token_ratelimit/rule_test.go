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
	"strings"
	"testing"
)

// Test helper functions to create test data
func createTestRuleForRule() *Rule {
	return &Rule{
		ID:       "test-rule-1",
		Resource: "test-resource",
		Strategy: PETA,
		Encoding: TokenEncoding{
			Provider: OpenAIEncoderProvider,
			Model:    "gpt-3.5-turbo",
		},
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{
					Type:  AllIdentifier,
					Value: ".*",
				},
				KeyItems: []*KeyItem{
					{
						Key: "default",
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
	}
}

func createEmptyRule() *Rule {
	return &Rule{}
}

func createRuleWithMultipleItems() *Rule {
	return &Rule{
		ID:       "multi-item-rule",
		Resource: "multi-resource",
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
						Key: "premium",
						Token: Token{
							Number:        5000,
							CountStrategy: TotalTokens,
						},
						Time: Time{
							Unit:  Hour,
							Value: 1,
						},
					},
					{
						Key: "basic",
						Token: Token{
							Number:        1000,
							CountStrategy: InputTokens,
						},
						Time: Time{
							Unit:  Minute,
							Value: 30,
						},
					},
				},
			},
			{
				Identifier: Identifier{
					Type:  AllIdentifier,
					Value: "fallback",
				},
				KeyItems: []*KeyItem{
					{
						Key: "default",
						Token: Token{
							Number:        500,
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
	}
}

// Test Rule.ResourceName function
func TestRule_ResourceName(t *testing.T) {
	tests := []struct {
		name     string
		rule     *Rule
		expected string
	}{
		{
			name:     "Valid rule with resource name",
			rule:     &Rule{Resource: "test-service"},
			expected: "test-service",
		},
		{
			name:     "Rule with empty resource name",
			rule:     &Rule{Resource: ""},
			expected: "",
		},
		{
			name:     "Nil rule",
			rule:     nil,
			expected: "Rule{nil}",
		},
		{
			name:     "Rule with special characters in resource",
			rule:     &Rule{Resource: "test-service-*-regex"},
			expected: "test-service-*-regex",
		},
		{
			name:     "Rule with spaces in resource",
			rule:     &Rule{Resource: "test service with spaces"},
			expected: "test service with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.ResourceName()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Test Rule.String function
func TestRule_String(t *testing.T) {
	tests := []struct {
		name        string
		rule        *Rule
		contains    []string // Expected substrings in the output
		notContains []string // Substrings that should not be in the output
	}{
		{
			name: "Complete rule with all fields",
			rule: createTestRuleForRule(),
			contains: []string{
				"Rule{",
				"ID:test-rule-1",
				"Resource:test-resource",
				"Strategy:peta",
				"Encoding:TokenEncoding{Provider:openai, Model:gpt-3.5-turbo}",
				"SpecificItems:[",
				"}",
			},
		},
		{
			name: "Rule without ID",
			rule: &Rule{
				Resource: "test-resource",
				Strategy: FixedWindow,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-4",
				},
				SpecificItems: []*SpecificItem{},
			},
			contains: []string{
				"Rule{",
				"Resource:test-resource",
				"Strategy:fixed-window",
				"SpecificItems:[]",
			},
			notContains: []string{
				"ID:",
			},
		},
		{
			name: "Rule with empty specific items",
			rule: &Rule{
				ID:            "empty-items",
				Resource:      "empty-resource",
				Strategy:      PETA,
				Encoding:      TokenEncoding{Provider: OpenAIEncoderProvider, Model: "gpt-3.5-turbo"},
				SpecificItems: []*SpecificItem{},
			},
			contains: []string{
				"SpecificItems:[]",
			},
		},
		{
			name:     "Nil rule",
			rule:     nil,
			contains: []string{"Rule{nil}"},
		},
		{
			name: "Rule with multiple specific items",
			rule: createRuleWithMultipleItems(),
			contains: []string{
				"Rule{",
				"ID:multi-item-rule",
				"Resource:multi-resource",
				"Strategy:fixed-window",
				"SpecificItems:[",
				", ", // Should contain comma separator between items
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.String()

			// Check that expected substrings are present
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected string to contain %q, got: %s", substr, result)
				}
			}

			// Check that unwanted substrings are not present
			for _, substr := range tt.notContains {
				if strings.Contains(result, substr) {
					t.Errorf("Expected string to NOT contain %q, got: %s", substr, result)
				}
			}
		})
	}
}

// Test Rule.setDefaultRuleOption function
func TestRule_SetDefaultRuleOption(t *testing.T) {
	tests := []struct {
		name         string
		rule         *Rule
		expectedRule *Rule
		description  string
	}{
		{
			name: "Empty rule should get default values",
			rule: &Rule{
				Encoding: TokenEncoding{Provider: OpenAIEncoderProvider},
			},
			expectedRule: &Rule{
				Resource: DefaultResourcePattern,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    DefaultTokenEncodingModel[OpenAIEncoderProvider],
				},
			},
			description: "Should set default resource pattern and model",
		},
		{
			name: "Rule with empty resource should get default",
			rule: &Rule{
				Resource: "",
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "custom-model",
				},
			},
			expectedRule: &Rule{
				Resource: DefaultResourcePattern,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "custom-model",
				},
			},
			description: "Should set default resource but keep custom model",
		},
		{
			name: "Rule with empty model should get default",
			rule: &Rule{
				Resource: "custom-resource",
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "",
				},
			},
			expectedRule: &Rule{
				Resource: "custom-resource",
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    DefaultTokenEncodingModel[OpenAIEncoderProvider],
				},
			},
			description: "Should keep custom resource but set default model",
		},
		{
			name: "Rule with specific items having empty values",
			rule: &Rule{
				Resource: "test-resource",
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-4",
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  AllIdentifier,
							Value: "",
						},
						KeyItems: []*KeyItem{
							{
								Key:   "",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
				},
			},
			expectedRule: &Rule{
				Resource: "test-resource",
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-4",
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  AllIdentifier,
							Value: DefaultIdentifierValuePattern,
						},
						KeyItems: []*KeyItem{
							{
								Key:   DefaultKeyPattern,
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
				},
			},
			description: "Should set default identifier value and key pattern",
		},
		{
			name: "Rule with multiple specific items and key items",
			rule: &Rule{
				Resource: "",
				Encoding: TokenEncoding{Provider: OpenAIEncoderProvider},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: ""},
						KeyItems: []*KeyItem{
							{Key: "", Token: Token{Number: 100}, Time: Time{Unit: Second, Value: 1}},
							{Key: "valid-key", Token: Token{Number: 200}, Time: Time{Unit: Minute, Value: 1}},
						},
					},
					{
						Identifier: Identifier{Type: Header, Value: "user-id"},
						KeyItems: []*KeyItem{
							{Key: "", Token: Token{Number: 300}, Time: Time{Unit: Hour, Value: 1}},
						},
					},
				},
			},
			expectedRule: &Rule{
				Resource: DefaultResourcePattern,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    DefaultTokenEncodingModel[OpenAIEncoderProvider],
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: DefaultIdentifierValuePattern},
						KeyItems: []*KeyItem{
							{Key: DefaultKeyPattern, Token: Token{Number: 100}, Time: Time{Unit: Second, Value: 1}},
							{Key: "valid-key", Token: Token{Number: 200}, Time: Time{Unit: Minute, Value: 1}},
						},
					},
					{
						Identifier: Identifier{Type: Header, Value: "user-id"},
						KeyItems: []*KeyItem{
							{Key: DefaultKeyPattern, Token: Token{Number: 300}, Time: Time{Unit: Hour, Value: 1}},
						},
					},
				},
			},
			description: "Should set defaults for multiple items correctly",
		},
		{
			name:         "Nil rule should not panic",
			rule:         nil,
			expectedRule: nil,
			description:  "Should handle nil rule gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the original rule to avoid modifying test data
			var testRule *Rule
			if tt.rule != nil {
				testRule = &Rule{
					ID:            tt.rule.ID,
					Resource:      tt.rule.Resource,
					Strategy:      tt.rule.Strategy,
					Encoding:      tt.rule.Encoding,
					SpecificItems: make([]*SpecificItem, len(tt.rule.SpecificItems)),
				}
				for i, item := range tt.rule.SpecificItems {
					if item != nil {
						testRule.SpecificItems[i] = &SpecificItem{
							Identifier: item.Identifier,
							KeyItems:   make([]*KeyItem, len(item.KeyItems)),
						}
						for j, keyItem := range item.KeyItems {
							if keyItem != nil {
								testRule.SpecificItems[i].KeyItems[j] = &KeyItem{
									Key:   keyItem.Key,
									Token: keyItem.Token,
									Time:  keyItem.Time,
								}
							}
						}
					}
				}
			}

			// Apply defaults
			testRule.setDefaultRuleOption()

			// Verify results
			if tt.expectedRule == nil {
				if testRule != nil {
					t.Error("Expected nil rule after setDefaultRuleOption")
				}
				return
			}

			if testRule == nil {
				t.Fatal("Expected non-nil rule after setDefaultRuleOption")
			}

			// Check basic fields
			if testRule.Resource != tt.expectedRule.Resource {
				t.Errorf("Resource: expected %q, got %q", tt.expectedRule.Resource, testRule.Resource)
			}

			if testRule.Encoding.Model != tt.expectedRule.Encoding.Model {
				t.Errorf("Model: expected %q, got %q", tt.expectedRule.Encoding.Model, testRule.Encoding.Model)
			}

			// Check specific items
			if len(testRule.SpecificItems) != len(tt.expectedRule.SpecificItems) {
				t.Errorf("SpecificItems length: expected %d, got %d",
					len(tt.expectedRule.SpecificItems), len(testRule.SpecificItems))
				return
			}

			for i, expectedItem := range tt.expectedRule.SpecificItems {
				actualItem := testRule.SpecificItems[i]

				if actualItem.Identifier.Value != expectedItem.Identifier.Value {
					t.Errorf("SpecificItem[%d].Identifier.Value: expected %q, got %q",
						i, expectedItem.Identifier.Value, actualItem.Identifier.Value)
				}

				if len(actualItem.KeyItems) != len(expectedItem.KeyItems) {
					t.Errorf("SpecificItem[%d].KeyItems length: expected %d, got %d",
						i, len(expectedItem.KeyItems), len(actualItem.KeyItems))
					continue
				}

				for j, expectedKeyItem := range expectedItem.KeyItems {
					actualKeyItem := actualItem.KeyItems[j]

					if actualKeyItem.Key != expectedKeyItem.Key {
						t.Errorf("SpecificItem[%d].KeyItems[%d].Key: expected %q, got %q",
							i, j, expectedKeyItem.Key, actualKeyItem.Key)
					}
				}
			}
		})
	}
}

// Test Rule.filterDuplicatedItem function
func TestRule_FilterDuplicatedItem(t *testing.T) {
	tests := []struct {
		name        string
		rule        *Rule
		expectedLen int
		description string
		verify      func(*Rule) bool
	}{
		{
			name: "Rule with duplicate key items",
			rule: &Rule{
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "test"},
						KeyItems: []*KeyItem{
							{
								Key:   "duplicate-key",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
							{
								Key:   "duplicate-key",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
							{
								Key:   "unique-key",
								Token: Token{Number: 500, CountStrategy: InputTokens},
								Time:  Time{Unit: Second, Value: 30},
							},
						},
					},
				},
			},
			expectedLen: 1,
			description: "Should remove duplicate key items",
			verify: func(r *Rule) bool {
				if len(r.SpecificItems) != 1 {
					return false
				}
				if len(r.SpecificItems[0].KeyItems) != 2 {
					return false
				}
				// Should keep the last occurrence (reverse order processing)
				return r.SpecificItems[0].KeyItems[0].Key == "unique-key"
			},
		},
		{
			name: "Rule with duplicate specific items",
			rule: &Rule{
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "same"},
						KeyItems: []*KeyItem{
							{
								Key:   "key1",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "same"},
						KeyItems: []*KeyItem{
							{
								Key:   "key1",
								Token: Token{Number: 500, CountStrategy: TotalTokens},
								Time:  Time{Unit: Second, Value: 60},
							},
						},
					},
				},
			},
			expectedLen: 1,
			description: "Should handle duplicate specific items",
			verify: func(r *Rule) bool {
				return len(r.SpecificItems) == 1 && len(r.SpecificItems[0].KeyItems) >= 1
			},
		},
		{
			name: "Rule with no duplicates",
			rule: &Rule{
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "test1"},
						KeyItems: []*KeyItem{
							{
								Key:   "key1",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
					{
						Identifier: Identifier{Type: Header, Value: "test2"},
						KeyItems: []*KeyItem{
							{
								Key:   "key2",
								Token: Token{Number: 500, CountStrategy: InputTokens},
								Time:  Time{Unit: Second, Value: 30},
							},
						},
					},
				},
			},
			expectedLen: 2,
			description: "Should preserve all unique items",
			verify: func(r *Rule) bool {
				return len(r.SpecificItems) == 2
			},
		},
		{
			name: "Rule with empty specific items",
			rule: &Rule{
				SpecificItems: []*SpecificItem{},
			},
			expectedLen: 0,
			description: "Should handle empty specific items",
			verify: func(r *Rule) bool {
				return len(r.SpecificItems) == 0
			},
		},
		{
			name: "Rule with specific item having empty key items",
			rule: &Rule{
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "test"},
						KeyItems:   []*KeyItem{},
					},
					{
						Identifier: Identifier{Type: Header, Value: "valid"},
						KeyItems: []*KeyItem{
							{
								Key:   "valid-key",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
				},
			},
			expectedLen: 1,
			description: "Should remove specific items with empty key items",
			verify: func(r *Rule) bool {
				return len(r.SpecificItems) == 1 &&
					r.SpecificItems[0].Identifier.Value == "valid"
			},
		},
		{
			name: "Complex rule with mixed duplicates",
			rule: &Rule{
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "id1"},
						KeyItems: []*KeyItem{
							{Key: "key1", Token: Token{Number: 100, CountStrategy: TotalTokens}, Time: Time{Unit: Second, Value: 1}},
							{Key: "key2", Token: Token{Number: 200, CountStrategy: InputTokens}, Time: Time{Unit: Minute, Value: 1}},
							{Key: "key1", Token: Token{Number: 100, CountStrategy: TotalTokens}, Time: Time{Unit: Second, Value: 1}}, // Duplicate
						},
					},
					{
						Identifier: Identifier{Type: Header, Value: "id2"},
						KeyItems: []*KeyItem{
							{Key: "key3", Token: Token{Number: 300, CountStrategy: OutputTokens}, Time: Time{Unit: Hour, Value: 1}},
						},
					},
					{
						Identifier: Identifier{Type: AllIdentifier, Value: "id1"},
						KeyItems: []*KeyItem{
							{Key: "key4", Token: Token{Number: 400, CountStrategy: TotalTokens}, Time: Time{Unit: Day, Value: 1}},
						},
					},
				},
			},
			expectedLen: 3,
			description: "Should handle complex duplicate scenarios",
			verify: func(r *Rule) bool {
				if len(r.SpecificItems) != 3 {
					return false
				}
				// Due to reverse processing, the order might be different
				// Just verify we have the expected identifiers
				hasId1 := false
				hasId2 := false
				hasKey4 := false
				hasKey1AndKey2 := false
				for _, item := range r.SpecificItems {
					if item.Identifier.Value == "id1" {
						hasId1 = true
						if len(item.KeyItems) == 1 {
							if item.KeyItems[0].Key == "key4" {
								hasKey4 = true
							}
						} else if len(item.KeyItems) == 2 {
							if item.KeyItems[0].Key == "key1" && item.KeyItems[1].Key == "key2" {
								hasKey1AndKey2 = true
							}
						} else {
							return false
						}
					}
					if item.Identifier.Value == "id2" {
						hasId2 = true
					}
				}
				return hasId1 && hasId2 && hasKey4 && hasKey1AndKey2
			},
		},
		{
			name:        "Nil rule should not panic",
			rule:        nil,
			expectedLen: 0,
			description: "Should handle nil rule gracefully",
			verify: func(r *Rule) bool {
				return r == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a deep copy to avoid modifying test data
			var testRule *Rule
			if tt.rule != nil {
				testRule = &Rule{
					ID:            tt.rule.ID,
					Resource:      tt.rule.Resource,
					Strategy:      tt.rule.Strategy,
					Encoding:      tt.rule.Encoding,
					SpecificItems: make([]*SpecificItem, len(tt.rule.SpecificItems)),
				}
				for i, item := range tt.rule.SpecificItems {
					if item != nil {
						testRule.SpecificItems[i] = &SpecificItem{
							Identifier: item.Identifier,
							KeyItems:   make([]*KeyItem, len(item.KeyItems)),
						}
						for j, keyItem := range item.KeyItems {
							if keyItem != nil {
								testRule.SpecificItems[i].KeyItems[j] = &KeyItem{
									Key:   keyItem.Key,
									Token: keyItem.Token,
									Time:  keyItem.Time,
								}
							}
						}
					}
				}
			}

			// Apply filter
			testRule.filterDuplicatedItem()

			// Basic verification
			if tt.rule != nil && testRule != nil {
				if len(testRule.SpecificItems) > tt.expectedLen {
					t.Errorf("Expected at most %d specific items, got %d",
						tt.expectedLen, len(testRule.SpecificItems))
				}
			}

			// Custom verification
			if tt.verify != nil && !tt.verify(testRule) {
				t.Errorf("Custom verification failed: %s", tt.description)
			}
		})
	}
}

// Test filterDuplicatedItem with edge cases
func TestRule_FilterDuplicatedItem_EdgeCases(t *testing.T) {
	t.Run("Rule with nil specific items", func(t *testing.T) {
		rule := &Rule{
			SpecificItems: []*SpecificItem{nil, nil},
		}

		// Should not panic
		rule.filterDuplicatedItem()

		// Should result in empty specific items
		if len(rule.SpecificItems) != 0 {
			t.Errorf("Expected empty specific items, got %d", len(rule.SpecificItems))
		}
	})

	t.Run("Rule with nil key items", func(t *testing.T) {
		rule := &Rule{
			SpecificItems: []*SpecificItem{
				{
					Identifier: Identifier{Type: AllIdentifier, Value: "test"},
					KeyItems:   []*KeyItem{nil, nil},
				},
			},
		}

		// Should not panic
		rule.filterDuplicatedItem()

		// Should remove the specific item with no valid key items
		if len(rule.SpecificItems) != 0 {
			t.Errorf("Expected no specific items, got %d", len(rule.SpecificItems))
		}
	})
}

// Test the hash generation consistency
func TestRule_FilterDuplicatedItem_HashConsistency(t *testing.T) {
	// Create two rules with identical key items
	rule1 := &Rule{
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{Type: AllIdentifier, Value: "test"},
				KeyItems: []*KeyItem{
					{
						Key:   "same-key",
						Token: Token{Number: 1000, CountStrategy: TotalTokens},
						Time:  Time{Unit: Minute, Value: 1},
					},
				},
			},
		},
	}

	rule2 := &Rule{
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{Type: AllIdentifier, Value: "test"},
				KeyItems: []*KeyItem{
					{
						Key:   "same-key",
						Token: Token{Number: 1000, CountStrategy: TotalTokens},
						Time:  Time{Unit: Minute, Value: 1},
					},
					{
						Key:   "same-key",
						Token: Token{Number: 1000, CountStrategy: TotalTokens},
						Time:  Time{Unit: Minute, Value: 1},
					},
				},
			},
		},
	}

	rule1.filterDuplicatedItem()
	rule2.filterDuplicatedItem()

	// Both should have the same result
	if len(rule1.SpecificItems) != len(rule2.SpecificItems) {
		t.Error("Hash consistency failed: different number of specific items")
	}

	if len(rule1.SpecificItems) > 0 && len(rule2.SpecificItems) > 0 {
		if len(rule1.SpecificItems[0].KeyItems) != len(rule2.SpecificItems[0].KeyItems) {
			t.Error("Hash consistency failed: different number of key items")
		}
	}
}

// Test integration between functions
func TestRule_Integration(t *testing.T) {
	// Test that setDefaultRuleOption and filterDuplicatedItem work together
	rule := &Rule{
		Encoding: TokenEncoding{Provider: OpenAIEncoderProvider},
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{Type: AllIdentifier, Value: ""},
				KeyItems: []*KeyItem{
					{Key: "", Token: Token{Number: 1000}, Time: Time{Unit: Minute, Value: 1}},
					{Key: "", Token: Token{Number: 1000}, Time: Time{Unit: Minute, Value: 1}},
				},
			},
		},
	}

	// Apply defaults first
	rule.setDefaultRuleOption()

	// Then filter duplicates
	rule.filterDuplicatedItem()

	// Verify the result
	if len(rule.SpecificItems) != 1 {
		t.Errorf("Expected 1 specific item, got %d", len(rule.SpecificItems))
	}

	if len(rule.SpecificItems[0].KeyItems) != 1 {
		t.Errorf("Expected 1 key item after filtering duplicates, got %d",
			len(rule.SpecificItems[0].KeyItems))
	}

	if rule.Resource != DefaultResourcePattern {
		t.Errorf("Expected default resource pattern %q, got %q",
			DefaultResourcePattern, rule.Resource)
	}

	if rule.SpecificItems[0].KeyItems[0].Key != DefaultKeyPattern {
		t.Errorf("Expected default key pattern %q, got %q",
			DefaultKeyPattern, rule.SpecificItems[0].KeyItems[0].Key)
	}
}

// Benchmark tests
func BenchmarkRule_ResourceName(b *testing.B) {
	rule := createTestRuleForRule()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule.ResourceName()
	}
}

func BenchmarkRule_String(b *testing.B) {
	rule := createRuleWithMultipleItems()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.String()
	}
}

func BenchmarkRule_SetDefaultRuleOption(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rule := createEmptyRule()
		rule.setDefaultRuleOption()
	}
}

func BenchmarkRule_FilterDuplicatedItem(b *testing.B) {
	// Create a rule with many duplicates for benchmarking
	rule := &Rule{
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{Type: AllIdentifier, Value: "test"},
				KeyItems:   make([]*KeyItem, 100),
			},
		},
	}

	// Fill with duplicate items
	for i := 0; i < 100; i++ {
		rule.SpecificItems[0].KeyItems[i] = &KeyItem{
			Key:   "duplicate-key",
			Token: Token{Number: 1000, CountStrategy: TotalTokens},
			Time:  Time{Unit: Minute, Value: 1},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh copy for each iteration
		testRule := &Rule{
			SpecificItems: []*SpecificItem{
				{
					Identifier: rule.SpecificItems[0].Identifier,
					KeyItems:   make([]*KeyItem, len(rule.SpecificItems[0].KeyItems)),
				},
			},
		}
		copy(testRule.SpecificItems[0].KeyItems, rule.SpecificItems[0].KeyItems)

		testRule.filterDuplicatedItem()
	}
}
