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
	"testing"
)

// Test NewDefaultTokenCalculator function
func TestNewDefaultTokenCalculator(t *testing.T) {
	// Test creating a new default token calculator
	calculator := NewDefaultTokenCalculator()

	// Verify the calculator is not nil
	if calculator == nil {
		t.Fatal("Expected non-nil TokenCalculatorManager, got nil")
	}

	// Verify the calculators map is initialized
	if calculator.calculators == nil {
		t.Fatal("Expected non-nil calculators map, got nil")
	}

	// Verify all expected calculators are present
	expectedStrategies := []CountStrategy{TotalTokens, InputTokens, OutputTokens}
	for _, strategy := range expectedStrategies {
		calc := calculator.getCalculator(strategy)
		if calc == nil {
			t.Errorf("Expected calculator for strategy %v, got nil", strategy)
		}
	}

	// Verify specific calculator types
	if _, ok := calculator.getCalculator(TotalTokens).(*TotalTokensCalculator); !ok {
		t.Error("Expected TotalTokensCalculator for TotalTokens strategy")
	}
	if _, ok := calculator.getCalculator(InputTokens).(*InputTokensCalculator); !ok {
		t.Error("Expected InputTokensCalculator for InputTokens strategy")
	}
	if _, ok := calculator.getCalculator(OutputTokens).(*OutputTokensCalculator); !ok {
		t.Error("Expected OutputTokensCalculator for OutputTokens strategy")
	}
}

// Test TokenCalculatorManager.getCalculator function
func TestTokenCalculatorManager_getCalculator(t *testing.T) {
	tests := []struct {
		name       string
		manager    *TokenCalculatorManager
		strategy   CountStrategy
		expectNil  bool
		expectType interface{}
	}{
		{
			name:       "Get TotalTokens calculator",
			manager:    NewDefaultTokenCalculator(),
			strategy:   TotalTokens,
			expectNil:  false,
			expectType: &TotalTokensCalculator{},
		},
		{
			name:       "Get InputTokens calculator",
			manager:    NewDefaultTokenCalculator(),
			strategy:   InputTokens,
			expectNil:  false,
			expectType: &InputTokensCalculator{},
		},
		{
			name:       "Get OutputTokens calculator",
			manager:    NewDefaultTokenCalculator(),
			strategy:   OutputTokens,
			expectNil:  false,
			expectType: &OutputTokensCalculator{},
		},
		{
			name:      "Get non-existent calculator",
			manager:   NewDefaultTokenCalculator(),
			strategy:  CountStrategy(999), // Invalid strategy
			expectNil: true,
		},
		{
			name:      "Nil manager",
			manager:   nil,
			strategy:  TotalTokens,
			expectNil: true,
		},
		{
			name: "Manager with nil calculators map",
			manager: &TokenCalculatorManager{
				calculators: nil,
			},
			strategy:  TotalTokens,
			expectNil: true,
		},
		{
			name: "Manager with empty calculators map",
			manager: &TokenCalculatorManager{
				calculators: make(map[CountStrategy]TokenCalculator),
			},
			strategy:  TotalTokens,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := tt.manager.getCalculator(tt.strategy)
			if tt.expectNil {
				if calc != nil {
					t.Errorf("Expected nil calculator, got %T", calc)
				}
			} else {
				if calc == nil {
					t.Errorf("Expected non-nil calculator, got nil")
				}
				// Check specific type if provided
				if tt.expectType != nil {
					expectedType := tt.expectType
					switch expectedType.(type) {
					case *TotalTokensCalculator:
						if _, ok := calc.(*TotalTokensCalculator); !ok {
							t.Errorf("Expected TotalTokensCalculator, got %T", calc)
						}
					case *InputTokensCalculator:
						if _, ok := calc.(*InputTokensCalculator); !ok {
							t.Errorf("Expected InputTokensCalculator, got %T", calc)
						}
					case *OutputTokensCalculator:
						if _, ok := calc.(*OutputTokensCalculator); !ok {
							t.Errorf("Expected OutputTokensCalculator, got %T", calc)
						}
					}
				}
			}
		})
	}
}

// Test InputTokensCalculator.Calculate function
func TestInputTokensCalculator_Calculate(t *testing.T) {
	tests := []struct {
		name     string
		calc     *InputTokensCalculator
		ctx      *Context
		infos    *UsedTokenInfos
		expected int
	}{
		{
			name: "Valid input tokens",
			calc: &InputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
			expected: 100,
		},
		{
			name: "Zero input tokens",
			calc: &InputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  0,
				OutputTokens: 50,
				TotalTokens:  50,
			},
			expected: 0,
		},
		{
			name: "Negative input tokens",
			calc: &InputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  -10,
				OutputTokens: 50,
				TotalTokens:  40,
			},
			expected: -10,
		},
		{
			name: "Large input tokens",
			calc: &InputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  999999,
				OutputTokens: 1,
				TotalTokens:  1000000,
			},
			expected: 999999,
		},
		{
			name:     "Nil calculator",
			calc:     nil,
			ctx:      NewContext(),
			infos:    &UsedTokenInfos{InputTokens: 100},
			expected: 0,
		},
		{
			name:     "Nil token infos",
			calc:     &InputTokensCalculator{},
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
		{
			name:     "Nil context (should still work)",
			calc:     &InputTokensCalculator{},
			ctx:      nil,
			infos:    &UsedTokenInfos{InputTokens: 100},
			expected: 100,
		},
		{
			name:     "Both calculator and infos are nil",
			calc:     nil,
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.calc.Calculate(tt.ctx, tt.infos)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test OutputTokensCalculator.Calculate function
func TestOutputTokensCalculator_Calculate(t *testing.T) {
	tests := []struct {
		name     string
		calc     *OutputTokensCalculator
		ctx      *Context
		infos    *UsedTokenInfos
		expected int
	}{
		{
			name: "Valid output tokens",
			calc: &OutputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 75,
				TotalTokens:  175,
			},
			expected: 75,
		},
		{
			name: "Zero output tokens",
			calc: &OutputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 0,
				TotalTokens:  100,
			},
			expected: 0,
		},
		{
			name: "Negative output tokens",
			calc: &OutputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: -25,
				TotalTokens:  75,
			},
			expected: -25,
		},
		{
			name: "Large output tokens",
			calc: &OutputTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  1,
				OutputTokens: 888888,
				TotalTokens:  888889,
			},
			expected: 888888,
		},
		{
			name:     "Nil calculator",
			calc:     nil,
			ctx:      NewContext(),
			infos:    &UsedTokenInfos{OutputTokens: 75},
			expected: 0,
		},
		{
			name:     "Nil token infos",
			calc:     &OutputTokensCalculator{},
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
		{
			name:     "Nil context (should still work)",
			calc:     &OutputTokensCalculator{},
			ctx:      nil,
			infos:    &UsedTokenInfos{OutputTokens: 75},
			expected: 75,
		},
		{
			name:     "Both calculator and infos are nil",
			calc:     nil,
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.calc.Calculate(tt.ctx, tt.infos)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test TotalTokensCalculator.Calculate function
func TestTotalTokensCalculator_Calculate(t *testing.T) {
	tests := []struct {
		name     string
		calc     *TotalTokensCalculator
		ctx      *Context
		infos    *UsedTokenInfos
		expected int
	}{
		{
			name: "Valid total tokens from TotalTokens field",
			calc: &TotalTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  200, // This value is used directly
			},
			expected: 200,
		},
		{
			name: "Zero total tokens",
			calc: &TotalTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  0,
			},
			expected: 0,
		},
		{
			name: "Negative total tokens",
			calc: &TotalTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  -30,
			},
			expected: -30,
		},
		{
			name: "Large total tokens",
			calc: &TotalTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  500000,
				OutputTokens: 300000,
				TotalTokens:  1000000,
			},
			expected: 1000000,
		},
		{
			name: "Total tokens field differs from sum of input and output",
			calc: &TotalTokensCalculator{},
			ctx:  NewContext(),
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  200, // Different from 100+50=150
			},
			expected: 200, // Uses TotalTokens field directly
		},
		{
			name:     "Nil calculator",
			calc:     nil,
			ctx:      NewContext(),
			infos:    &UsedTokenInfos{TotalTokens: 150},
			expected: 0,
		},
		{
			name:     "Nil token infos",
			calc:     &TotalTokensCalculator{},
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
		{
			name: "Nil context (should still work)",
			calc: &TotalTokensCalculator{},
			ctx:  nil,
			infos: &UsedTokenInfos{
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
			expected: 150,
		},
		{
			name:     "Both calculator and infos are nil",
			calc:     nil,
			ctx:      NewContext(),
			infos:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.calc.Calculate(tt.ctx, tt.infos)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test integration with TokenCalculatorManager
func TestTokenCalculatorManager_Integration(t *testing.T) {
	manager := NewDefaultTokenCalculator()
	ctx := NewContext()
	infos := &UsedTokenInfos{
		InputTokens:  100,
		OutputTokens: 75,
		TotalTokens:  200, // Different from sum to test TotalTokens behavior
	}

	tests := []struct {
		name     string
		strategy CountStrategy
		expected int
	}{
		{
			name:     "Calculate input tokens",
			strategy: InputTokens,
			expected: 100,
		},
		{
			name:     "Calculate output tokens",
			strategy: OutputTokens,
			expected: 75,
		},
		{
			name:     "Calculate total tokens (uses TotalTokens field)",
			strategy: TotalTokens,
			expected: 200, // Uses TotalTokens field, not sum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := manager.getCalculator(tt.strategy)
			if calc == nil {
				t.Fatalf("Failed to get calculator for strategy %v", tt.strategy)
			}

			result := calc.Calculate(ctx, infos)
			if result != tt.expected {
				t.Errorf("Strategy %v: expected %d, got %d", tt.strategy, tt.expected, result)
			}
		})
	}
}

// Test edge cases with extreme values
func TestTokenCalculators_ExtremeValues(t *testing.T) {
	ctx := NewContext()

	tests := []struct {
		name       string
		calculator TokenCalculator
		infos      *UsedTokenInfos
		expected   int
	}{
		{
			name:       "InputTokensCalculator with max int",
			calculator: &InputTokensCalculator{},
			infos: &UsedTokenInfos{
				InputTokens:  2147483647, // Max int32
				OutputTokens: 0,
				TotalTokens:  2147483647,
			},
			expected: 2147483647,
		},
		{
			name:       "OutputTokensCalculator with min int",
			calculator: &OutputTokensCalculator{},
			infos: &UsedTokenInfos{
				InputTokens:  0,
				OutputTokens: -2147483648, // Min int32
				TotalTokens:  -2147483648,
			},
			expected: -2147483648,
		},
		{
			name:       "TotalTokensCalculator with zero",
			calculator: &TotalTokensCalculator{},
			infos: &UsedTokenInfos{
				InputTokens:  1000,
				OutputTokens: 500,
				TotalTokens:  0, // Zero total despite non-zero components
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.calculator.Calculate(ctx, tt.infos)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test context interactions (though current implementation doesn't use context)
func TestTokenCalculators_ContextInteraction(t *testing.T) {
	tests := []struct {
		name string
		ctx  *Context
	}{
		{
			name: "Empty context",
			ctx:  NewContext(),
		},
		{
			name: "Context with data",
			ctx: func() *Context {
				ctx := NewContext()
				ctx.Set("test-key", "test-value")
				return ctx
			}(),
		},
		{
			name: "Nil context",
			ctx:  nil,
		},
	}

	infos := &UsedTokenInfos{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	calculators := []struct {
		name string
		calc TokenCalculator
	}{
		{"InputTokensCalculator", &InputTokensCalculator{}},
		{"OutputTokensCalculator", &OutputTokensCalculator{}},
		{"TotalTokensCalculator", &TotalTokensCalculator{}},
	}

	for _, ctxTest := range tests {
		for _, calcTest := range calculators {
			t.Run(ctxTest.name+"_"+calcTest.name, func(t *testing.T) {
				// Should not panic and should return consistent results
				result := calcTest.calc.Calculate(ctxTest.ctx, infos)

				// Verify expected results based on calculator type
				switch calcTest.calc.(type) {
				case *InputTokensCalculator:
					if result != 100 {
						t.Errorf("Expected 100, got %d", result)
					}
				case *OutputTokensCalculator:
					if result != 50 {
						t.Errorf("Expected 50, got %d", result)
					}
				case *TotalTokensCalculator:
					if result != 150 {
						t.Errorf("Expected 150, got %d", result)
					}
				}
			})
		}
	}
}

// Test global token calculator variable
func TestGlobalTokenCalculator(t *testing.T) {
	// Test that global variable is initialized
	if globalTokenCalculator == nil {
		t.Fatal("Expected globalTokenCalculator to be initialized, got nil")
	}

	// Test that global calculator functions correctly
	calc := globalTokenCalculator.getCalculator(InputTokens)
	if calc == nil {
		t.Fatal("Expected to get InputTokensCalculator from global instance, got nil")
	}

	// Test calculation with global instance
	ctx := NewContext()
	infos := &UsedTokenInfos{InputTokens: 42}
	result := calc.Calculate(ctx, infos)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// Benchmark tests for performance
func BenchmarkTokenCalculators(b *testing.B) {
	ctx := NewContext()
	infos := &UsedTokenInfos{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
	}

	calculators := map[string]TokenCalculator{
		"InputTokens":  &InputTokensCalculator{},
		"OutputTokens": &OutputTokensCalculator{},
		"TotalTokens":  &TotalTokensCalculator{},
	}

	for name, calc := range calculators {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				calc.Calculate(ctx, infos)
			}
		})
	}
}

// Benchmark TokenCalculatorManager getCalculator
func BenchmarkTokenCalculatorManager_getCalculator(b *testing.B) {
	manager := NewDefaultTokenCalculator()
	strategies := []CountStrategy{InputTokens, OutputTokens, TotalTokens}

	for _, strategy := range strategies {
		b.Run(strategy.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				manager.getCalculator(strategy)
			}
		})
	}
}

// Benchmark complete calculation flow
func BenchmarkTokenCalculatorManager_CompleteFlow(b *testing.B) {
	manager := NewDefaultTokenCalculator()
	ctx := NewContext()
	infos := &UsedTokenInfos{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
	}

	strategies := []CountStrategy{InputTokens, OutputTokens, TotalTokens}

	for _, strategy := range strategies {
		b.Run(strategy.String()+"_CompleteFlow", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				calc := manager.getCalculator(strategy)
				if calc != nil {
					calc.Calculate(ctx, infos)
				}
			}
		})
	}
}
