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
	"testing"
	"time"
)

func TestBaseRuleCollector_Collect(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *Context
		rule     *Rule
		expected int // Expected number of returned rules
		wantNil  bool
	}{
		{
			name:    "nil collector",
			ctx:     NewContext(),
			rule:    &Rule{},
			wantNil: true,
		},
		{
			name:     "nil rule",
			ctx:      NewContext(),
			rule:     nil,
			expected: 0,
		},
		{
			name: "rule with nil SpecificItems",
			ctx:  NewContext(),
			rule: &Rule{
				Resource:      "test-resource",
				Strategy:      PETA,
				SpecificItems: nil,
			},
			expected: 0,
		},
		{
			name: "valid rule with header identifier",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"user123"},
					},
				})
				return ctx
			}(),
			rule: &Rule{
				Resource: "test-resource",
				Strategy: PETA,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-3.5-turbo",
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "User-Id",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user123",
								Token: Token{
									Number:        1000,
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  Second,
									Value: 60,
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "rule with multiple key items",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"user123"},
						"App-Id":  {"app456"},
					},
				})
				return ctx
			}(),
			rule: &Rule{
				Resource: "test-resource",
				Strategy: PETA,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-3.5-turbo",
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "User-Id",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user123",
								Token: Token{
									Number:        1000,
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  Second,
									Value: 60,
								},
							},
							{
								Key: "user123",
								Token: Token{
									Number:        500,
									CountStrategy: InputTokens,
								},
								Time: Time{
									Unit:  Minute,
									Value: 5,
								},
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "rule with duplicate limit keys (should deduplicate)",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"user123"},
					},
				})
				return ctx
			}(),
			rule: &Rule{
				Resource: "test-resource",
				Strategy: PETA,
				Encoding: TokenEncoding{
					Provider: OpenAIEncoderProvider,
					Model:    "gpt-3.5-turbo",
				},
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "User-Id",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user123",
								Token: Token{
									Number:        1000,
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  Second,
									Value: 60,
								},
							},
							{
								Key: "user123",
								Token: Token{
									Number:        1000, // Same configuration, should be deduplicated
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  Second,
									Value: 60,
								},
							},
						},
					},
				},
			},
			expected: 1, // Should be deduplicated, only return 1
		},
		{
			name: "rule with invalid time window",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"user123"},
					},
				})
				return ctx
			}(),
			rule: &Rule{
				Resource: "test-resource",
				Strategy: PETA,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "User-Id",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user123",
								Token: Token{
									Number:        1000,
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  TimeUnit(999), // Invalid time unit
									Value: 60,
								},
							},
						},
					},
				},
			},
			expected: 0, // Invalid time window, should be skipped
		},
		{
			name: "rule with unmatched identifier",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set(KeyRequestInfos, &RequestInfos{
					Headers: map[string][]string{
						"User-Id": {"user123"},
					},
				})
				return ctx
			}(),
			rule: &Rule{
				Resource: "test-resource",
				Strategy: PETA,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "User-Id",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user456", // Unmatched user ID
								Token: Token{
									Number:        1000,
									CountStrategy: TotalTokens,
								},
								Time: Time{
									Unit:  Second,
									Value: 60,
								},
							},
						},
					},
				},
			},
			expected: 0, // Identifier not matched, should return 0 rules
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var collector *BaseRuleCollector
			if tt.wantNil {
				collector = nil
			} else {
				collector = &BaseRuleCollector{}
			}

			result := collector.Collect(tt.ctx, tt.rule)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
				return
			}

			if len(result) != tt.expected {
				t.Errorf("Expected %d rules, got %d", tt.expected, len(result))
			}

			// Validate basic fields of returned rules
			for _, rule := range result {
				if rule.LimitKey == "" {
					t.Error("LimitKey should not be empty")
				}
				if rule.TimeWindow <= 0 {
					t.Error("TimeWindow should be positive")
				}
				if rule.TokenSize <= 0 {
					t.Error("TokenSize should be positive")
				}
			}
		})
	}
}

// Concurrency safety test
func TestBaseRuleCollector_Collect_Concurrency(t *testing.T) {
	collector := &BaseRuleCollector{}

	ctx := NewContext()
	ctx.Set(KeyRequestInfos, &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	})

	rule := &Rule{
		Resource: "test-resource",
		Strategy: PETA,
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{
					Type:  Header,
					Value: "User-Id",
				},
				KeyItems: []*KeyItem{
					{
						Key: "user123",
						Token: Token{
							Number:        1000,
							CountStrategy: TotalTokens,
						},
						Time: Time{
							Unit:  Second,
							Value: 60,
						},
					},
				},
			},
		},
	}

	// Execute concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				result := collector.Collect(ctx, rule)
				if len(result) != 1 {
					t.Errorf("Expected 1 rule, got %d", len(result))
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timeout")
		}
	}
}

// Edge cases test
func TestBaseRuleCollector_Collect_EdgeCases(t *testing.T) {
	collector := &BaseRuleCollector{}

	tests := []struct {
		name     string
		ctx      *Context
		rule     *Rule
		expected int
	}{
		{
			name: "empty specific items",
			ctx:  NewContext(),
			rule: &Rule{
				Resource:      "test",
				Strategy:      PETA,
				SpecificItems: []*SpecificItem{},
			},
			expected: 0,
		},
		{
			name: "specific item with empty key items",
			ctx:  NewContext(),
			rule: &Rule{
				Resource: "test",
				Strategy: PETA,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier},
						KeyItems:   []*KeyItem{},
					},
				},
			},
			expected: 0,
		},
		{
			name: "specific item with nil key items",
			ctx:  NewContext(),
			rule: &Rule{
				Resource: "test",
				Strategy: PETA,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: AllIdentifier},
						KeyItems:   nil,
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.Collect(tt.ctx, tt.rule)
			if len(result) != tt.expected {
				t.Errorf("Expected %d rules, got %d", tt.expected, len(result))
			}
		})
	}
}

// Performance test
func BenchmarkBaseRuleCollector_Collect(b *testing.B) {
	collector := &BaseRuleCollector{}

	// Prepare test data
	ctx := NewContext()
	ctx.Set(KeyRequestInfos, &RequestInfos{
		Headers: map[string][]string{
			"User-Id":  {"user123"},
			"App-Id":   {"app456"},
			"Version":  {"v1.0"},
			"Platform": {"web"},
		},
	})

	rule := &Rule{
		Resource: "test-resource",
		Strategy: PETA,
		Encoding: TokenEncoding{
			Provider: OpenAIEncoderProvider,
			Model:    "gpt-3.5-turbo",
		},
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{
					Type:  Header,
					Value: "User-Id",
				},
				KeyItems: []*KeyItem{
					{
						Key: "user123",
						Token: Token{
							Number:        1000,
							CountStrategy: TotalTokens,
						},
						Time: Time{
							Unit:  Second,
							Value: 60,
						},
					},
					{
						Key: "user123",
						Token: Token{
							Number:        500,
							CountStrategy: InputTokens,
						},
						Time: Time{
							Unit:  Minute,
							Value: 5,
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := collector.Collect(ctx, rule)
		_ = result // Avoid compiler optimization
	}
}

// Large rules performance test
func BenchmarkBaseRuleCollector_Collect_LargeRules(b *testing.B) {
	collector := &BaseRuleCollector{}

	// Prepare large amount of test data
	ctx := NewContext()
	headers := make(map[string][]string)
	for i := 0; i < 100; i++ {
		headers[fmt.Sprintf("Header-%d", i)] = []string{fmt.Sprintf("value-%d", i)}
	}
	ctx.Set(KeyRequestInfos, &RequestInfos{Headers: headers})

	// Create Rule with large amount of rules
	var specificItems []*SpecificItem
	for i := 0; i < 50; i++ {
		var keyItems []*KeyItem
		for j := 0; j < 10; j++ {
			keyItems = append(keyItems, &KeyItem{
				Key: fmt.Sprintf("value-%d", i),
				Token: Token{
					Number:        int64(1000 + j),
					CountStrategy: CountStrategy(j % 3),
				},
				Time: Time{
					Unit:  TimeUnit(j % 4),
					Value: int64(60 + j),
				},
			})
		}
		specificItems = append(specificItems, &SpecificItem{
			Identifier: Identifier{
				Type:  Header,
				Value: fmt.Sprintf("Header-%d", i),
			},
			KeyItems: keyItems,
		})
	}

	rule := &Rule{
		Resource:      "large-test-resource",
		Strategy:      PETA,
		SpecificItems: specificItems,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := collector.Collect(ctx, rule)
		_ = result
	}
}

// Memory allocation test
func BenchmarkBaseRuleCollector_Collect_Memory(b *testing.B) {
	collector := &BaseRuleCollector{}

	ctx := NewContext()
	ctx.Set(KeyRequestInfos, &RequestInfos{
		Headers: map[string][]string{
			"User-Id": {"user123"},
		},
	})

	rule := &Rule{
		Resource: "test-resource",
		Strategy: PETA,
		SpecificItems: []*SpecificItem{
			{
				Identifier: Identifier{
					Type:  Header,
					Value: "User-Id",
				},
				KeyItems: []*KeyItem{
					{
						Key: "user123",
						Token: Token{
							Number:        1000,
							CountStrategy: TotalTokens,
						},
						Time: Time{
							Unit:  Second,
							Value: 60,
						},
					},
				},
			},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := collector.Collect(ctx, rule)
		_ = result
	}
}
