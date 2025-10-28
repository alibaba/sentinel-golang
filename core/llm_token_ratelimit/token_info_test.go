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
	"testing"
)

// Test UsedTokenInfos struct initialization
func TestUsedTokenInfos_Initialization(t *testing.T) {
	// Test zero value initialization
	var infos UsedTokenInfos
	if infos.InputTokens != 0 {
		t.Errorf("Expected InputTokens to be 0, got %d", infos.InputTokens)
	}
	if infos.OutputTokens != 0 {
		t.Errorf("Expected OutputTokens to be 0, got %d", infos.OutputTokens)
	}
	if infos.TotalTokens != 0 {
		t.Errorf("Expected TotalTokens to be 0, got %d", infos.TotalTokens)
	}

	// Test struct literal initialization
	infos = UsedTokenInfos{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}
	if infos.InputTokens != 100 {
		t.Errorf("Expected InputTokens to be 100, got %d", infos.InputTokens)
	}
	if infos.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens to be 50, got %d", infos.OutputTokens)
	}
	if infos.TotalTokens != 150 {
		t.Errorf("Expected TotalTokens to be 150, got %d", infos.TotalTokens)
	}
}

// Test WithInputTokens function
func TestWithInputTokens(t *testing.T) {
	tests := []struct {
		name        string
		inputTokens int
	}{
		{
			name:        "Positive input tokens",
			inputTokens: 100,
		},
		{
			name:        "Zero input tokens",
			inputTokens: 0,
		},
		{
			name:        "Negative input tokens",
			inputTokens: -50,
		},
		{
			name:        "Large input tokens",
			inputTokens: 999999,
		},
		{
			name:        "Maximum integer value",
			inputTokens: 2147483647,
		},
		{
			name:        "Minimum integer value",
			inputTokens: -2147483648,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the function
			tokenInfoFunc := WithInputTokens(tt.inputTokens)

			// Verify the function is not nil
			if tokenInfoFunc == nil {
				t.Fatal("Expected non-nil UsedTokenInfo function, got nil")
			}

			// Test applying the function to a UsedTokenInfos struct
			infos := &UsedTokenInfos{}
			tokenInfoFunc(infos)

			// Verify the InputTokens field is set correctly
			if infos.InputTokens != tt.inputTokens {
				t.Errorf("Expected InputTokens to be %d, got %d", tt.inputTokens, infos.InputTokens)
			}

			// Verify other fields remain unchanged (zero values)
			if infos.OutputTokens != 0 {
				t.Errorf("Expected OutputTokens to remain 0, got %d", infos.OutputTokens)
			}
			if infos.TotalTokens != 0 {
				t.Errorf("Expected TotalTokens to remain 0, got %d", infos.TotalTokens)
			}
		})
	}
}

// Test WithOutputTokens function
func TestWithOutputTokens(t *testing.T) {
	tests := []struct {
		name         string
		outputTokens int
	}{
		{
			name:         "Positive output tokens",
			outputTokens: 75,
		},
		{
			name:         "Zero output tokens",
			outputTokens: 0,
		},
		{
			name:         "Negative output tokens",
			outputTokens: -25,
		},
		{
			name:         "Large output tokens",
			outputTokens: 888888,
		},
		{
			name:         "Maximum integer value",
			outputTokens: 2147483647,
		},
		{
			name:         "Minimum integer value",
			outputTokens: -2147483648,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the function
			tokenInfoFunc := WithOutputTokens(tt.outputTokens)

			// Verify the function is not nil
			if tokenInfoFunc == nil {
				t.Fatal("Expected non-nil UsedTokenInfo function, got nil")
			}

			// Test applying the function to a UsedTokenInfos struct
			infos := &UsedTokenInfos{}
			tokenInfoFunc(infos)

			// Verify the OutputTokens field is set correctly
			if infos.OutputTokens != tt.outputTokens {
				t.Errorf("Expected OutputTokens to be %d, got %d", tt.outputTokens, infos.OutputTokens)
			}

			// Verify other fields remain unchanged (zero values)
			if infos.InputTokens != 0 {
				t.Errorf("Expected InputTokens to remain 0, got %d", infos.InputTokens)
			}
			if infos.TotalTokens != 0 {
				t.Errorf("Expected TotalTokens to remain 0, got %d", infos.TotalTokens)
			}
		})
	}
}

// Test WithTotalTokens function
func TestWithTotalTokens(t *testing.T) {
	tests := []struct {
		name        string
		totalTokens int
	}{
		{
			name:        "Positive total tokens",
			totalTokens: 150,
		},
		{
			name:        "Zero total tokens",
			totalTokens: 0,
		},
		{
			name:        "Negative total tokens",
			totalTokens: -100,
		},
		{
			name:        "Large total tokens",
			totalTokens: 1000000,
		},
		{
			name:        "Maximum integer value",
			totalTokens: 2147483647,
		},
		{
			name:        "Minimum integer value",
			totalTokens: -2147483648,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the function
			tokenInfoFunc := WithTotalTokens(tt.totalTokens)

			// Verify the function is not nil
			if tokenInfoFunc == nil {
				t.Fatal("Expected non-nil UsedTokenInfo function, got nil")
			}

			// Test applying the function to a UsedTokenInfos struct
			infos := &UsedTokenInfos{}
			tokenInfoFunc(infos)

			// Verify the TotalTokens field is set correctly
			if infos.TotalTokens != tt.totalTokens {
				t.Errorf("Expected TotalTokens to be %d, got %d", tt.totalTokens, infos.TotalTokens)
			}

			// Verify other fields remain unchanged (zero values)
			if infos.InputTokens != 0 {
				t.Errorf("Expected InputTokens to remain 0, got %d", infos.InputTokens)
			}
			if infos.OutputTokens != 0 {
				t.Errorf("Expected OutputTokens to remain 0, got %d", infos.OutputTokens)
			}
		})
	}
}

// Test GenerateUsedTokenInfos function with no arguments
func TestGenerateUsedTokenInfos_NoArguments(t *testing.T) {
	// Test with no arguments
	infos := GenerateUsedTokenInfos()

	// Should return a valid pointer
	if infos == nil {
		t.Fatal("Expected non-nil UsedTokenInfos, got nil")
	}

	// All fields should be zero values
	if infos.InputTokens != 0 {
		t.Errorf("Expected InputTokens to be 0, got %d", infos.InputTokens)
	}
	if infos.OutputTokens != 0 {
		t.Errorf("Expected OutputTokens to be 0, got %d", infos.OutputTokens)
	}
	if infos.TotalTokens != 0 {
		t.Errorf("Expected TotalTokens to be 0, got %d", infos.TotalTokens)
	}
}

// Test GenerateUsedTokenInfos function with single arguments
func TestGenerateUsedTokenInfos_SingleArguments(t *testing.T) {
	tests := []struct {
		name     string
		function UsedTokenInfo
		expected UsedTokenInfos
	}{
		{
			name:     "Only input tokens",
			function: WithInputTokens(100),
			expected: UsedTokenInfos{InputTokens: 100, OutputTokens: 0, TotalTokens: 0},
		},
		{
			name:     "Only output tokens",
			function: WithOutputTokens(50),
			expected: UsedTokenInfos{InputTokens: 0, OutputTokens: 50, TotalTokens: 0},
		},
		{
			name:     "Only total tokens",
			function: WithTotalTokens(150),
			expected: UsedTokenInfos{InputTokens: 0, OutputTokens: 0, TotalTokens: 150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := GenerateUsedTokenInfos(tt.function)

			if infos == nil {
				t.Fatal("Expected non-nil UsedTokenInfos, got nil")
			}

			if infos.InputTokens != tt.expected.InputTokens {
				t.Errorf("Expected InputTokens to be %d, got %d", tt.expected.InputTokens, infos.InputTokens)
			}
			if infos.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("Expected OutputTokens to be %d, got %d", tt.expected.OutputTokens, infos.OutputTokens)
			}
			if infos.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("Expected TotalTokens to be %d, got %d", tt.expected.TotalTokens, infos.TotalTokens)
			}
		})
	}
}

// Test GenerateUsedTokenInfos function with multiple arguments
func TestGenerateUsedTokenInfos_MultipleArguments(t *testing.T) {
	tests := []struct {
		name      string
		functions []UsedTokenInfo
		expected  UsedTokenInfos
	}{
		{
			name: "All token types",
			functions: []UsedTokenInfo{
				WithInputTokens(100),
				WithOutputTokens(50),
				WithTotalTokens(150),
			},
			expected: UsedTokenInfos{InputTokens: 100, OutputTokens: 50, TotalTokens: 150},
		},
		{
			name: "Overlapping input tokens (last one wins)",
			functions: []UsedTokenInfo{
				WithInputTokens(100),
				WithInputTokens(200),
				WithOutputTokens(50),
			},
			expected: UsedTokenInfos{InputTokens: 200, OutputTokens: 50, TotalTokens: 0},
		},
		{
			name: "Mixed order",
			functions: []UsedTokenInfo{
				WithTotalTokens(300),
				WithInputTokens(200),
				WithOutputTokens(100),
			},
			expected: UsedTokenInfos{InputTokens: 200, OutputTokens: 100, TotalTokens: 300},
		},
		{
			name: "Duplicate functions",
			functions: []UsedTokenInfo{
				WithInputTokens(100),
				WithOutputTokens(50),
				WithInputTokens(150), // This should override the first one
				WithTotalTokens(200),
			},
			expected: UsedTokenInfos{InputTokens: 150, OutputTokens: 50, TotalTokens: 200},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := GenerateUsedTokenInfos(tt.functions...)

			if infos == nil {
				t.Fatal("Expected non-nil UsedTokenInfos, got nil")
			}

			if infos.InputTokens != tt.expected.InputTokens {
				t.Errorf("Expected InputTokens to be %d, got %d", tt.expected.InputTokens, infos.InputTokens)
			}
			if infos.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("Expected OutputTokens to be %d, got %d", tt.expected.OutputTokens, infos.OutputTokens)
			}
			if infos.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("Expected TotalTokens to be %d, got %d", tt.expected.TotalTokens, infos.TotalTokens)
			}
		})
	}
}

// Test GenerateUsedTokenInfos function with nil functions
func TestGenerateUsedTokenInfos_WithNilFunctions(t *testing.T) {
	tests := []struct {
		name      string
		functions []UsedTokenInfo
		expected  UsedTokenInfos
	}{
		{
			name:      "Single nil function",
			functions: []UsedTokenInfo{nil},
			expected:  UsedTokenInfos{InputTokens: 0, OutputTokens: 0, TotalTokens: 0},
		},
		{
			name: "Mixed nil and valid functions",
			functions: []UsedTokenInfo{
				WithInputTokens(100),
				nil,
				WithOutputTokens(50),
				nil,
				WithTotalTokens(150),
			},
			expected: UsedTokenInfos{InputTokens: 100, OutputTokens: 50, TotalTokens: 150},
		},
		{
			name:      "All nil functions",
			functions: []UsedTokenInfo{nil, nil, nil},
			expected:  UsedTokenInfos{InputTokens: 0, OutputTokens: 0, TotalTokens: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := GenerateUsedTokenInfos(tt.functions...)

			if infos == nil {
				t.Fatal("Expected non-nil UsedTokenInfos, got nil")
			}

			if infos.InputTokens != tt.expected.InputTokens {
				t.Errorf("Expected InputTokens to be %d, got %d", tt.expected.InputTokens, infos.InputTokens)
			}
			if infos.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("Expected OutputTokens to be %d, got %d", tt.expected.OutputTokens, infos.OutputTokens)
			}
			if infos.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("Expected TotalTokens to be %d, got %d", tt.expected.TotalTokens, infos.TotalTokens)
			}
		})
	}
}

// Test extractUsedTokenInfos function with nil context
func TestExtractUsedTokenInfos_NilContext(t *testing.T) {
	result := extractUsedTokenInfos(nil)

	if result != nil {
		t.Errorf("Expected nil result for nil context, got %v", result)
	}
}

// Test extractUsedTokenInfos function with empty context
func TestExtractUsedTokenInfos_EmptyContext(t *testing.T) {
	ctx := NewContext()
	result := extractUsedTokenInfos(ctx)

	if result != nil {
		t.Errorf("Expected nil result for empty context, got %v", result)
	}
}

// Test extractUsedTokenInfos function with valid token infos
func TestExtractUsedTokenInfos_ValidTokenInfos(t *testing.T) {
	ctx := NewContext()
	expectedInfos := &UsedTokenInfos{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	// Set the token infos in context
	ctx.Set(KeyUsedTokenInfos, expectedInfos)

	result := extractUsedTokenInfos(ctx)

	if result == nil {
		t.Fatal("Expected non-nil result, got nil")
	}

	if result.InputTokens != expectedInfos.InputTokens {
		t.Errorf("Expected InputTokens to be %d, got %d", expectedInfos.InputTokens, result.InputTokens)
	}
	if result.OutputTokens != expectedInfos.OutputTokens {
		t.Errorf("Expected OutputTokens to be %d, got %d", expectedInfos.OutputTokens, result.OutputTokens)
	}
	if result.TotalTokens != expectedInfos.TotalTokens {
		t.Errorf("Expected TotalTokens to be %d, got %d", expectedInfos.TotalTokens, result.TotalTokens)
	}
}

// Test extractUsedTokenInfos function with invalid data types
func TestExtractUsedTokenInfos_InvalidDataTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "String value",
			value: "not token infos",
		},
		{
			name:  "Integer value",
			value: 42,
		},
		{
			name:  "Map value",
			value: map[string]int{"tokens": 100},
		},
		{
			name:  "Slice value",
			value: []int{100, 50, 150},
		},
		{
			name:  "Boolean value",
			value: true,
		},
		{
			name:  "Nil pointer to UsedTokenInfos",
			value: (*UsedTokenInfos)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set(KeyUsedTokenInfos, tt.value)

			result := extractUsedTokenInfos(ctx)

			if result != nil {
				t.Errorf("Expected nil result for invalid data type, got %v", result)
			}
		})
	}
}

// Test extractUsedTokenInfos function with different context states
func TestExtractUsedTokenInfos_ContextStates(t *testing.T) {
	// Test with context that has other data but not token infos
	t.Run("Context with other data", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("other_key", "other_value")
		ctx.Set("another_key", 123)

		result := extractUsedTokenInfos(ctx)

		if result != nil {
			t.Errorf("Expected nil result when token infos not present, got %v", result)
		}
	})

	// Test with context that has token infos and other data
	t.Run("Context with token infos and other data", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("other_key", "other_value")

		expectedInfos := &UsedTokenInfos{
			InputTokens:  200,
			OutputTokens: 100,
			TotalTokens:  300,
		}
		ctx.Set(KeyUsedTokenInfos, expectedInfos)
		ctx.Set("another_key", 456)

		result := extractUsedTokenInfos(ctx)

		if result == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if result.InputTokens != expectedInfos.InputTokens {
			t.Errorf("Expected InputTokens to be %d, got %d", expectedInfos.InputTokens, result.InputTokens)
		}
		if result.OutputTokens != expectedInfos.OutputTokens {
			t.Errorf("Expected OutputTokens to be %d, got %d", expectedInfos.OutputTokens, result.OutputTokens)
		}
		if result.TotalTokens != expectedInfos.TotalTokens {
			t.Errorf("Expected TotalTokens to be %d, got %d", expectedInfos.TotalTokens, result.TotalTokens)
		}
	})
}

// Test integration between all functions
func TestTokenInfo_Integration(t *testing.T) {
	// Create token infos using the builder pattern
	infos := GenerateUsedTokenInfos(
		WithInputTokens(500),
		WithOutputTokens(300),
		WithTotalTokens(800),
	)

	// Store in context
	ctx := NewContext()
	ctx.Set(KeyUsedTokenInfos, infos)

	// Extract from context
	extracted := extractUsedTokenInfos(ctx)

	// Verify they match
	if extracted == nil {
		t.Fatal("Expected non-nil extracted infos, got nil")
	}

	if extracted.InputTokens != 500 {
		t.Errorf("Expected InputTokens to be 500, got %d", extracted.InputTokens)
	}
	if extracted.OutputTokens != 300 {
		t.Errorf("Expected OutputTokens to be 300, got %d", extracted.OutputTokens)
	}
	if extracted.TotalTokens != 800 {
		t.Errorf("Expected TotalTokens to be 800, got %d", extracted.TotalTokens)
	}

	// Verify they are the same instance
	if extracted != infos {
		t.Error("Expected extracted infos to be the same instance as original")
	}
}

// Benchmark tests for performance
func BenchmarkWithInputTokens(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn := WithInputTokens(100)
		infos := &UsedTokenInfos{}
		fn(infos)
	}
}

func BenchmarkWithOutputTokens(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn := WithOutputTokens(50)
		infos := &UsedTokenInfos{}
		fn(infos)
	}
}

func BenchmarkWithTotalTokens(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn := WithTotalTokens(150)
		infos := &UsedTokenInfos{}
		fn(infos)
	}
}

func BenchmarkGenerateUsedTokenInfos(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateUsedTokenInfos(
			WithInputTokens(100),
			WithOutputTokens(50),
			WithTotalTokens(150),
		)
	}
}

func BenchmarkExtractUsedTokenInfos(b *testing.B) {
	ctx := NewContext()
	tokenInfos := &UsedTokenInfos{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}
	ctx.Set(KeyUsedTokenInfos, tokenInfos)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractUsedTokenInfos(ctx)
	}
}
