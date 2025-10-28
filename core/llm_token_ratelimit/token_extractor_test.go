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

// Test OpenAITokenExtractor function with valid responses
func TestOpenAITokenExtractor_ValidResponses(t *testing.T) {
	tests := []struct {
		name     string
		response interface{}
		expected *UsedTokenInfos
	}{
		{
			name: "Standard OpenAI response with all tokens",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			expected: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
		},
		{
			name: "Response with zero tokens",
			response: map[string]any{
				"prompt_tokens":     0,
				"completion_tokens": 0,
				"total_tokens":      0,
			},
			expected: &UsedTokenInfos{
				InputTokens:  0,
				OutputTokens: 0,
				TotalTokens:  0,
			},
		},
		{
			name: "Response with large token counts",
			response: map[string]any{
				"prompt_tokens":     10000,
				"completion_tokens": 5000,
				"total_tokens":      15000,
			},
			expected: &UsedTokenInfos{
				InputTokens:  10000,
				OutputTokens: 5000,
				TotalTokens:  15000,
			},
		},
		{
			name: "Response with mixed token values",
			response: map[string]any{
				"prompt_tokens":     1,
				"completion_tokens": 999,
				"total_tokens":      1000,
			},
			expected: &UsedTokenInfos{
				InputTokens:  1,
				OutputTokens: 999,
				TotalTokens:  1000,
			},
		},
		{
			name: "Response with additional fields (should be ignored)",
			response: map[string]any{
				"prompt_tokens":     200,
				"completion_tokens": 100,
				"total_tokens":      300,
				"model":             "gpt-3.5-turbo",
				"id":                "chatcmpl-123",
				"object":            "chat.completion",
			},
			expected: &UsedTokenInfos{
				InputTokens:  200,
				OutputTokens: 100,
				TotalTokens:  300,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result, got nil")
			}

			if result.InputTokens != tt.expected.InputTokens {
				t.Errorf("InputTokens: expected %d, got %d", tt.expected.InputTokens, result.InputTokens)
			}
			if result.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("OutputTokens: expected %d, got %d", tt.expected.OutputTokens, result.OutputTokens)
			}
			if result.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("TotalTokens: expected %d, got %d", tt.expected.TotalTokens, result.TotalTokens)
			}
		})
	}
}

// Test OpenAITokenExtractor function with nil input
func TestOpenAITokenExtractor_NilInput(t *testing.T) {
	result, err := OpenAITokenExtractor(nil)

	// Should return error for nil input
	if err == nil {
		t.Error("Expected error for nil input, got nil")
	}

	// Should return nil result
	if result != nil {
		t.Errorf("Expected nil result for nil input, got %v", result)
	}

	// Check error message
	expectedErrorMsg := "response is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// Test OpenAITokenExtractor function with invalid response types
func TestOpenAITokenExtractor_InvalidResponseTypes(t *testing.T) {
	tests := []struct {
		name          string
		response      interface{}
		expectedError string
	}{
		{
			name:          "String response",
			response:      "not a map",
			expectedError: "response is not map[string]any",
		},
		{
			name:          "Integer response",
			response:      42,
			expectedError: "response is not map[string]any",
		},
		{
			name:          "Slice response",
			response:      []string{"test"},
			expectedError: "response is not map[string]any",
		},
		{
			name:          "Boolean response",
			response:      true,
			expectedError: "response is not map[string]any",
		},
		{
			name:          "Float response",
			response:      3.14,
			expectedError: "response is not map[string]any",
		},
		{
			name: "Map with wrong key type",
			response: map[int]any{
				1: 100,
				2: 50,
				3: 150,
			},
			expectedError: "response is not map[string]any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)

			// Should return error
			if err == nil {
				t.Error("Expected error, got nil")
			}

			// Should return nil result
			if result != nil {
				t.Errorf("Expected nil result, got %v", result)
			}

			// Check error message
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// Test OpenAITokenExtractor function with missing required fields
func TestOpenAITokenExtractor_MissingFields(t *testing.T) {
	tests := []struct {
		name          string
		response      map[string]any
		expectedError string
	}{
		{
			name: "Missing prompt_tokens",
			response: map[string]any{
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			expectedError: "prompt_tokens not found or not int",
		},
		{
			name: "Missing completion_tokens",
			response: map[string]any{
				"prompt_tokens": 100,
				"total_tokens":  150,
			},
			expectedError: "completion_tokens not found or not int",
		},
		{
			name: "Missing total_tokens",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
			},
			expectedError: "total_tokens not found or not int",
		},
		{
			name:          "Missing all token fields",
			response:      map[string]any{},
			expectedError: "prompt_tokens not found or not int",
		},
		{
			name: "Empty map",
			response: map[string]any{
				"other_field": "value",
			},
			expectedError: "prompt_tokens not found or not int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)

			// Should return error
			if err == nil {
				t.Error("Expected error, got nil")
			}

			// Should return nil result
			if result != nil {
				t.Errorf("Expected nil result, got %v", result)
			}

			// Check error message
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// Test OpenAITokenExtractor function with invalid token value types
func TestOpenAITokenExtractor_InvalidTokenTypes(t *testing.T) {
	tests := []struct {
		name          string
		response      map[string]any
		expectedError string
	}{
		{
			name: "prompt_tokens as string",
			response: map[string]any{
				"prompt_tokens":     "100",
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			expectedError: "prompt_tokens not found or not int",
		},
		{
			name: "completion_tokens as float",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50.5,
				"total_tokens":      150,
			},
			expectedError: "completion_tokens not found or not int",
		},
		{
			name: "total_tokens as boolean",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      true,
			},
			expectedError: "total_tokens not found or not int",
		},
		{
			name: "prompt_tokens as nil",
			response: map[string]any{
				"prompt_tokens":     nil,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			expectedError: "prompt_tokens not found or not int",
		},
		{
			name: "completion_tokens as slice",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": []int{50},
				"total_tokens":      150,
			},
			expectedError: "completion_tokens not found or not int",
		},
		{
			name: "total_tokens as map",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      map[string]int{"value": 150},
			},
			expectedError: "total_tokens not found or not int",
		},
		{
			name: "All tokens as wrong types",
			response: map[string]any{
				"prompt_tokens":     "100",
				"completion_tokens": 50.0,
				"total_tokens":      false,
			},
			expectedError: "prompt_tokens not found or not int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)

			// Should return error
			if err == nil {
				t.Error("Expected error, got nil")
			}

			// Should return nil result
			if result != nil {
				t.Errorf("Expected nil result, got %v", result)
			}

			// Check error message
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// Test OpenAITokenExtractor function with negative token values
func TestOpenAITokenExtractor_NegativeTokenValues(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]any
		expected *UsedTokenInfos
	}{
		{
			name: "Negative prompt_tokens",
			response: map[string]any{
				"prompt_tokens":     -10,
				"completion_tokens": 50,
				"total_tokens":      40,
			},
			expected: &UsedTokenInfos{
				InputTokens:  -10,
				OutputTokens: 50,
				TotalTokens:  40,
			},
		},
		{
			name: "Negative completion_tokens",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": -25,
				"total_tokens":      75,
			},
			expected: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: -25,
				TotalTokens:  75,
			},
		},
		{
			name: "Negative total_tokens",
			response: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      -10,
			},
			expected: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  -10,
			},
		},
		{
			name: "All negative values",
			response: map[string]any{
				"prompt_tokens":     -100,
				"completion_tokens": -50,
				"total_tokens":      -150,
			},
			expected: &UsedTokenInfos{
				InputTokens:  -100,
				OutputTokens: -50,
				TotalTokens:  -150,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)

			// Should not return error for negative values
			if err != nil {
				t.Errorf("Expected no error for negative values, got %v", err)
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result, got nil")
			}

			if result.InputTokens != tt.expected.InputTokens {
				t.Errorf("InputTokens: expected %d, got %d", tt.expected.InputTokens, result.InputTokens)
			}
			if result.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("OutputTokens: expected %d, got %d", tt.expected.OutputTokens, result.OutputTokens)
			}
			if result.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("TotalTokens: expected %d, got %d", tt.expected.TotalTokens, result.TotalTokens)
			}
		})
	}
}

// Test OpenAITokenExtractor function with extreme integer values
func TestOpenAITokenExtractor_ExtremeValues(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]any
		expected *UsedTokenInfos
	}{
		{
			name: "Maximum integer values",
			response: map[string]any{
				"prompt_tokens":     2147483647, // max int32
				"completion_tokens": 2147483647,
				"total_tokens":      2147483647,
			},
			expected: &UsedTokenInfos{
				InputTokens:  2147483647,
				OutputTokens: 2147483647,
				TotalTokens:  2147483647,
			},
		},
		{
			name: "Minimum integer values",
			response: map[string]any{
				"prompt_tokens":     -2147483648, // min int32
				"completion_tokens": -2147483648,
				"total_tokens":      -2147483648,
			},
			expected: &UsedTokenInfos{
				InputTokens:  -2147483648,
				OutputTokens: -2147483648,
				TotalTokens:  -2147483648,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpenAITokenExtractor(tt.response)

			if err != nil {
				t.Errorf("Expected no error for extreme values, got %v", err)
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result, got nil")
			}

			if result.InputTokens != tt.expected.InputTokens {
				t.Errorf("InputTokens: expected %d, got %d", tt.expected.InputTokens, result.InputTokens)
			}
			if result.OutputTokens != tt.expected.OutputTokens {
				t.Errorf("OutputTokens: expected %d, got %d", tt.expected.OutputTokens, result.OutputTokens)
			}
			if result.TotalTokens != tt.expected.TotalTokens {
				t.Errorf("TotalTokens: expected %d, got %d", tt.expected.TotalTokens, result.TotalTokens)
			}
		})
	}
}

// Test OpenAITokenExtractor integration with GenerateUsedTokenInfos
func TestOpenAITokenExtractor_Integration(t *testing.T) {
	response := map[string]any{
		"prompt_tokens":     123,
		"completion_tokens": 456,
		"total_tokens":      579,
	}

	result, err := OpenAITokenExtractor(response)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the result uses GenerateUsedTokenInfos correctly
	expected := GenerateUsedTokenInfos(
		WithInputTokens(123),
		WithOutputTokens(456),
		WithTotalTokens(579),
	)

	if result.InputTokens != expected.InputTokens {
		t.Errorf("InputTokens: expected %d, got %d", expected.InputTokens, result.InputTokens)
	}
	if result.OutputTokens != expected.OutputTokens {
		t.Errorf("OutputTokens: expected %d, got %d", expected.OutputTokens, result.OutputTokens)
	}
	if result.TotalTokens != expected.TotalTokens {
		t.Errorf("TotalTokens: expected %d, got %d", expected.TotalTokens, result.TotalTokens)
	}
}

// Benchmark OpenAITokenExtractor performance
func BenchmarkOpenAITokenExtractor(b *testing.B) {
	response := map[string]any{
		"prompt_tokens":     1000,
		"completion_tokens": 500,
		"total_tokens":      1500,
		"model":             "gpt-3.5-turbo",
		"id":                "chatcmpl-123",
		"object":            "chat.completion",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := OpenAITokenExtractor(response)
		if err != nil {
			b.Fatalf("Unexpected error in benchmark: %v", err)
		}
	}
}

// Benchmark OpenAITokenExtractor with error cases
func BenchmarkOpenAITokenExtractor_ErrorCases(b *testing.B) {
	errorCases := []struct {
		name     string
		response interface{}
	}{
		{"nil_response", nil},
		{"invalid_type", "not a map"},
		{"missing_fields", map[string]any{"other": "value"}},
		{"wrong_field_type", map[string]any{
			"prompt_tokens":     "100",
			"completion_tokens": 50,
			"total_tokens":      150,
		}},
	}

	for _, ec := range errorCases {
		b.Run(ec.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = OpenAITokenExtractor(ec.response)
			}
		})
	}
}
