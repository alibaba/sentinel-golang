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
	"reflect"
	"strings"
	"testing"
)

// Test helper functions to create test data
func createTestTokenEncoding(provider TokenEncoderProvider, model string) TokenEncoding {
	return TokenEncoding{
		Provider: provider,
		Model:    model,
	}
}

func createTestContextForEncoder() *Context {
	ctx := NewContext()
	ctx.Set(KeyRequestID, "test-request-123")
	return ctx
}

func createTestMatchedRuleForEncoder() *MatchedRule {
	return &MatchedRule{
		Strategy:      PETA,
		LimitKey:      "test-limit-key",
		TimeWindow:    60, // 60 seconds
		TokenSize:     1000,
		CountStrategy: TotalTokens,
		// PETA specific fields
		Encoding: TokenEncoding{
			Provider: OpenAIEncoderProvider,
			Model:    "gpt-3.5-turbo",
		},
		EstimatedToken: 100,
	}
}

func resetTokenEncoderGlobalState() {
	// Clear the global token encoder map for clean testing
	tokenEncoderMapRWMux.Lock()
	defer tokenEncoderMapRWMux.Unlock()
	tokenEncoderMap = make(map[TokenEncoding]TokenEncoder)
}

// Test NewTokenEncoder function
func TestNewTokenEncoder(t *testing.T) {
	defer resetTokenEncoderGlobalState()

	tests := []struct {
		name        string
		ctx         *Context
		encoding    TokenEncoding
		description string
	}{
		{
			name:        "OpenAI provider should create OpenAIEncoder",
			ctx:         createTestContextForEncoder(),
			encoding:    createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			description: "Should create OpenAIEncoder for OpenAI provider",
		},
		{
			name:        "OpenAI provider with different model",
			ctx:         createTestContextForEncoder(),
			encoding:    createTestTokenEncoding(OpenAIEncoderProvider, "gpt-4"),
			description: "Should create OpenAIEncoder for different OpenAI model",
		},
		{
			name:        "Unknown provider should fallback to OpenAIEncoder",
			ctx:         createTestContextForEncoder(),
			encoding:    TokenEncoding{Provider: TokenEncoderProvider(999), Model: "unknown-model"},
			description: "Should fallback to OpenAIEncoder for unknown provider",
		},
		{
			name:        "Empty model should use default",
			ctx:         createTestContextForEncoder(),
			encoding:    createTestTokenEncoding(OpenAIEncoderProvider, ""),
			description: "Should handle empty model gracefully",
		},
		{
			name:        "Nil context should not panic",
			ctx:         nil,
			encoding:    createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			description: "Should handle nil context gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new encoder
			encoder := NewTokenEncoder(tt.ctx, tt.encoding)

			// Verify encoder is not nil
			if encoder == nil {
				t.Error("Expected non-nil encoder")
				return
			}

			if openAIEncoder, ok := encoder.(*OpenAIEncoder); ok {
				if openAIEncoder.Model == "" {
					t.Error("Expected OpenAIEncoder to have a model set")
				}
			} else {
				t.Error("Expected encoder to be OpenAIEncoder type")
			}

			// Verify encoder is stored in global map
			storedEncoder := LookupTokenEncoder(tt.ctx, tt.encoding)
			if storedEncoder == nil {
				t.Error("Expected encoder to be stored in global map")
			}

			if !reflect.DeepEqual(storedEncoder, encoder) {
				t.Error("Expected stored encoder to be the same instance")
			}
		})
	}
}

// Test LookupTokenEncoder function
func TestLookupTokenEncoder(t *testing.T) {
	defer resetTokenEncoderGlobalState()

	tests := []struct {
		name        string
		ctx         *Context
		encoding    TokenEncoding
		setupFunc   func(TokenEncoding) TokenEncoder
		expectNil   bool
		description string
	}{
		{
			name:     "Existing encoder should be returned",
			ctx:      createTestContextForEncoder(),
			encoding: createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			setupFunc: func(encoding TokenEncoding) TokenEncoder {
				return NewTokenEncoder(createTestContextForEncoder(), encoding)
			},
			expectNil:   false,
			description: "Should return existing encoder from map",
		},
		{
			name:        "Non-existing encoder should return nil",
			ctx:         createTestContextForEncoder(),
			encoding:    createTestTokenEncoding(OpenAIEncoderProvider, "non-existing-model"),
			setupFunc:   nil,
			expectNil:   true,
			description: "Should return nil for non-existing encoder",
		},
		{
			name:     "Different encodings should return different encoders",
			ctx:      createTestContextForEncoder(),
			encoding: createTestTokenEncoding(OpenAIEncoderProvider, "gpt-4"),
			setupFunc: func(encoding TokenEncoding) TokenEncoder {
				// Create encoder for gpt-3.5-turbo first
				NewTokenEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"))
				// Then create for the test encoding
				return NewTokenEncoder(createTestContextForEncoder(), encoding)
			},
			expectNil:   false,
			description: "Should handle multiple different encoders",
		},
		{
			name:     "Nil context should not affect lookup",
			ctx:      nil,
			encoding: createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			setupFunc: func(encoding TokenEncoding) TokenEncoder {
				// Create encoder for gpt-3.5-turbo first
				NewTokenEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"))
				// Then create for the test encoding
				return NewTokenEncoder(createTestContextForEncoder(), encoding)
			},
			expectNil:   false,
			description: "Should handle nil context in lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup encoder if needed
			var expectedEncoder TokenEncoder
			if tt.setupFunc != nil {
				expectedEncoder = tt.setupFunc(tt.encoding)
			}

			// Lookup encoder
			encoder := LookupTokenEncoder(tt.ctx, tt.encoding)

			// Verify result
			if tt.expectNil {
				if encoder != nil {
					t.Error("Expected nil encoder, got non-nil")
				}
			} else {
				if encoder == nil {
					t.Error("Expected non-nil encoder, got nil")
				}
				if expectedEncoder != nil && !reflect.DeepEqual(expectedEncoder, encoder) {
					t.Error("Expected returned encoder to match the created encoder")
				}
			}
		})
	}
}

// Test NewOpenAIEncoder function
func TestNewOpenAIEncoder(t *testing.T) {
	tests := []struct {
		name         string
		ctx          *Context
		encoding     TokenEncoding
		expectModel  string
		expectNilEnc bool
		description  string
	}{
		{
			name:         "Valid model should create encoder successfully",
			ctx:          createTestContextForEncoder(),
			encoding:     createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			expectModel:  "gpt-3.5-turbo",
			expectNilEnc: false,
			description:  "Should create encoder with correct model",
		},
		{
			name:         "GPT-4 model should work",
			ctx:          createTestContextForEncoder(),
			encoding:     createTestTokenEncoding(OpenAIEncoderProvider, "gpt-4"),
			expectModel:  "gpt-4",
			expectNilEnc: false,
			description:  "Should create encoder for GPT-4 model",
		},
		{
			name:         "Invalid model should fallback to default",
			ctx:          createTestContextForEncoder(),
			encoding:     createTestTokenEncoding(OpenAIEncoderProvider, "invalid-model-12345"),
			expectModel:  DefaultTokenEncodingModel[OpenAIEncoderProvider],
			expectNilEnc: false,
			description:  "Should fallback to default model for invalid model",
		},
		{
			name:         "Empty model should fallback to default",
			ctx:          createTestContextForEncoder(),
			encoding:     createTestTokenEncoding(OpenAIEncoderProvider, ""),
			expectModel:  DefaultTokenEncodingModel[OpenAIEncoderProvider],
			expectNilEnc: false,
			description:  "Should fallback to default model for empty model",
		},
		{
			name:         "Nil context should not panic",
			ctx:          nil,
			encoding:     createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"),
			expectModel:  "gpt-3.5-turbo",
			expectNilEnc: false,
			description:  "Should handle nil context gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create encoder
			encoder := NewOpenAIEncoder(tt.ctx, tt.encoding)

			// Verify encoder is not nil
			if encoder == nil {
				t.Error("Expected non-nil OpenAIEncoder")
				return
			}

			// Verify model
			if encoder.Model != tt.expectModel {
				t.Errorf("Expected model %s, got %s", tt.expectModel, encoder.Model)
			}

			// Verify internal tiktoken encoder
			if tt.expectNilEnc {
				if encoder.Encoder != nil {
					t.Error("Expected nil internal encoder")
				}
			} else {
				if encoder.Encoder == nil {
					t.Error("Expected non-nil internal encoder")
				}
			}
		})
	}
}

// Test OpenAIEncoder.CountTokens function
func TestOpenAIEncoder_CountTokens(t *testing.T) {
	tests := []struct {
		name        string
		encoder     *OpenAIEncoder
		ctx         *Context
		prompts     []string
		rule        *MatchedRule
		expectCount int
		expectError bool
		description string
	}{
		{
			name:        "Valid encoder with simple prompts",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"Hello", "World"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // We can't predict exact count, just verify > 0
			expectError: false,
			description: "Should count tokens for valid prompts",
		},
		{
			name:        "Empty prompts should return zero",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: 0,
			expectError: false,
			description: "Should return 0 for empty prompts",
		},
		{
			name:        "Single prompt should work",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"How are you?"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should count tokens for single prompt",
		},
		{
			name:        "Long prompt should work",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"This is a very long prompt that contains many words and should result in multiple tokens being generated by the tokenizer"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should count tokens for long prompt",
		},
		{
			name:        "Multiple prompts should concatenate",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"First prompt.", "Second prompt.", "Third prompt."},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should concatenate and count tokens for multiple prompts",
		},
		{
			name:        "Nil encoder should return error",
			encoder:     nil,
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"Hello"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: 0,
			expectError: true,
			description: "Should return error for nil encoder",
		},
		{
			name: "Encoder with nil internal encoder should return error",
			encoder: &OpenAIEncoder{
				Model:   "test-model",
				Encoder: nil,
			},
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"Hello"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: 0,
			expectError: true,
			description: "Should return error for encoder with nil internal encoder",
		},
		{
			name:        "Nil context should not affect counting",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         nil,
			prompts:     []string{"Hello"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should handle nil context gracefully",
		},
		{
			name:        "Nil rule should not affect counting",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"Hello"},
			rule:        nil,
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should handle nil rule gracefully",
		},
		{
			name:        "Empty strings in prompts should work",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"", "Hello", ""},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should handle empty strings in prompts",
		},
		{
			name:        "Special characters should work",
			encoder:     NewOpenAIEncoder(createTestContextForEncoder(), createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")),
			ctx:         createTestContextForEncoder(),
			prompts:     []string{"Hello! @#$%^&*()"},
			rule:        createTestMatchedRuleForEncoder(),
			expectCount: -1, // Verify > 0
			expectError: false,
			description: "Should handle special characters in prompts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count tokens
			count, err := tt.encoder.CountTokens(tt.ctx, tt.prompts, tt.rule)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if count != tt.expectCount {
					t.Errorf("Expected count %d when error occurs, got %d", tt.expectCount, count)
				}
				return
			}

			// Verify no error when not expected
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}

			// Verify count
			if tt.expectCount >= 0 {
				if count != tt.expectCount {
					t.Errorf("Expected count %d, got %d", tt.expectCount, count)
				}
			} else {
				// For cases where we expect > 0 but can't predict exact count
				if len(tt.prompts) > 0 && strings.Join(tt.prompts, "") != "" {
					if count <= 0 {
						t.Errorf("Expected positive token count for non-empty prompts, got %d", count)
					}
				}
			}
		})
	}
}

// Test OpenAIEncoder.CountTokens with different models
func TestOpenAIEncoder_CountTokens_DifferentModels(t *testing.T) {
	testPrompts := []string{"Hello, how are you today?"}
	ctx := createTestContextForEncoder()
	rule := createTestMatchedRuleForEncoder()

	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"text-davinci-003", // If supported
	}

	for _, model := range models {
		t.Run("Model_"+model, func(t *testing.T) {
			encoder := NewOpenAIEncoder(ctx, createTestTokenEncoding(OpenAIEncoderProvider, model))
			if encoder == nil {
				t.Error("Expected non-nil encoder")
				return
			}

			count, err := encoder.CountTokens(ctx, testPrompts, rule)
			if err != nil {
				t.Errorf("Expected no error for model %s, got: %v", model, err)
				return
			}

			if count <= 0 {
				t.Errorf("Expected positive token count for model %s, got %d", model, count)
			}

			t.Logf("Model %s token count: %d", encoder.Model, count)
		})
	}
}

// Test concurrent access to token encoder map
func TestTokenEncoder_ConcurrentAccess(t *testing.T) {
	defer resetTokenEncoderGlobalState()

	const numGoroutines = 10
	const numOperations = 20

	encoding := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")
	ctx := createTestContextForEncoder()

	// Channel to collect results
	results := make(chan TokenEncoder, numGoroutines*numOperations)
	errors := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numOperations; j++ {
				// Alternate between creating and looking up encoders
				if j%2 == 0 {
					encoder := NewTokenEncoder(ctx, encoding)
					results <- encoder
				} else {
					encoder := LookupTokenEncoder(ctx, encoding)
					results <- encoder
				}
			}
		}()
	}

	// Collect results
	encoderCount := 0
	nilCount := 0
	for i := 0; i < numGoroutines*numOperations; i++ {
		select {
		case encoder := <-results:
			if encoder != nil {
				encoderCount++
			} else {
				nilCount++
			}
		case err := <-errors:
			t.Errorf("Unexpected error in concurrent access: %v", err)
		}
	}

	// Verify we got some non-nil encoders
	if encoderCount == 0 {
		t.Error("Expected some non-nil encoders from concurrent operations")
	}

	// Verify final state
	finalEncoder := LookupTokenEncoder(ctx, encoding)
	if finalEncoder == nil {
		t.Error("Expected final lookup to return non-nil encoder")
	}
}

// Test edge cases and error conditions
func TestTokenEncoder_EdgeCases(t *testing.T) {
	defer resetTokenEncoderGlobalState()

	t.Run("Multiple encodings should coexist", func(t *testing.T) {
		ctx := createTestContextForEncoder()

		encoding1 := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")
		encoding2 := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-4")

		encoder1 := NewTokenEncoder(ctx, encoding1)
		encoder2 := NewTokenEncoder(ctx, encoding2)

		if encoder1 == nil || encoder2 == nil {
			t.Error("Expected both encoders to be non-nil")
		}

		if reflect.DeepEqual(encoder1, encoder2) {
			t.Error("Expected different encoders for different encodings")
		}

		// Verify both can be looked up
		lookup1 := LookupTokenEncoder(ctx, encoding1)
		lookup2 := LookupTokenEncoder(ctx, encoding2)

		if !reflect.DeepEqual(lookup1, encoder1) || !reflect.DeepEqual(lookup2, encoder2) {
			t.Error("Expected lookups to return original encoders")
		}
	})

	t.Run("TokenEncoder interface compliance", func(t *testing.T) {
		ctx := createTestContextForEncoder()
		encoding := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")

		encoder := NewTokenEncoder(ctx, encoding)

		// Verify it implements TokenEncoder interface
		var _ TokenEncoder = encoder

		// Test the interface method
		count, err := encoder.CountTokens(ctx, []string{"test"}, createTestMatchedRuleForEncoder())
		if err != nil {
			t.Errorf("Expected no error from interface method, got: %v", err)
		}
		if count <= 0 {
			t.Error("Expected positive token count from interface method")
		}
	})
}

// Benchmark tests
func BenchmarkNewTokenEncoder(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoding := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTokenEncoder(ctx, encoding)
	}
}

func BenchmarkLookupTokenEncoder(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoding := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")

	// Setup encoder
	NewTokenEncoder(ctx, encoding)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LookupTokenEncoder(ctx, encoding)
	}
}

func BenchmarkNewOpenAIEncoder(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoding := createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewOpenAIEncoder(ctx, encoding)
	}
}

func BenchmarkOpenAIEncoder_CountTokens(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoder := NewOpenAIEncoder(ctx, createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"))
	prompts := []string{"This is a test prompt for benchmarking token counting performance."}
	rule := createTestMatchedRuleForEncoder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.CountTokens(ctx, prompts, rule)
	}
}

func BenchmarkOpenAIEncoder_CountTokens_LongPrompt(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoder := NewOpenAIEncoder(ctx, createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"))
	longPrompt := strings.Repeat("This is a long prompt with many repeated words for testing performance. ", 100)
	prompts := []string{longPrompt}
	rule := createTestMatchedRuleForEncoder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.CountTokens(ctx, prompts, rule)
	}
}

func BenchmarkOpenAIEncoder_CountTokens_MultiplePrompts(b *testing.B) {
	ctx := createTestContextForEncoder()
	encoder := NewOpenAIEncoder(ctx, createTestTokenEncoding(OpenAIEncoderProvider, "gpt-3.5-turbo"))
	prompts := []string{
		"First prompt for testing.",
		"Second prompt for testing.",
		"Third prompt for testing.",
		"Fourth prompt for testing.",
		"Fifth prompt for testing.",
	}
	rule := createTestMatchedRuleForEncoder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.CountTokens(ctx, prompts, rule)
	}
}
