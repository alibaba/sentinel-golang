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

func TestIsValidRule(t *testing.T) {
	tests := []struct {
		name      string
		rule      *Rule
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil rule",
			rule:      nil,
			wantError: true,
			errorMsg:  "rule cannot be nil",
		},
		{
			name: "valid rule",
			rule: &Rule{
				Resource: "test-resource",
				Strategy: FixedWindow,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{
							Type:  Header,
							Value: "api-key",
						},
						KeyItems: []*KeyItem{
							{
								Key: "user-*",
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
			wantError: false,
		},
		{
			name: "empty rule items",
			rule: &Rule{
				Resource:      "test-resource",
				Strategy:      FixedWindow,
				SpecificItems: []*SpecificItem{},
			},
			wantError: true,
			errorMsg:  "specificItems cannot be empty",
		},
		{
			name: "invalid resource",
			rule: &Rule{
				Resource: "",
				Strategy: FixedWindow,
				SpecificItems: []*SpecificItem{
					{
						Identifier: Identifier{Type: Header, Value: "api-key"},
						KeyItems: []*KeyItem{
							{
								Key:   "user-*",
								Token: Token{Number: 1000, CountStrategy: TotalTokens},
								Time:  Time{Unit: Minute, Value: 1},
							},
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "invalid resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidRule(tt.rule)
			if tt.wantError {
				if err == nil {
					t.Errorf("IsValidRule() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("IsValidRule() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("IsValidRule() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateResource(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty resource",
			resource:  "",
			wantError: true,
			errorMsg:  "resource pattern cannot be empty",
		},
		{
			name:      "valid resource pattern",
			resource:  "api-.*",
			wantError: false,
		},
		{
			name:      "valid wildcard pattern",
			resource:  ".*",
			wantError: false,
		},
		{
			name:      "invalid regex pattern",
			resource:  "[invalid",
			wantError: true,
			errorMsg:  "resource pattern is not a valid regex",
		},
		{
			name:      "complex valid pattern",
			resource:  "^(api|service)-.+$",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResource(tt.resource)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateResource() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateResource() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateResource() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateStrategy(t *testing.T) {
	tests := []struct {
		name      string
		strategy  Strategy
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid fixed window strategy",
			strategy:  FixedWindow,
			wantError: false,
		},
		{
			name:      "invalid strategy",
			strategy:  Strategy(999),
			wantError: true,
			errorMsg:  "unsupported strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStrategy(tt.strategy)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateStrategy() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateStrategy() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateStrategy() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier *Identifier
		wantError  bool
		errorMsg   string
	}{
		{
			name:       "nil identifier",
			identifier: nil,
			wantError:  true,
			errorMsg:   "identifier cannot be nil",
		},
		{
			name: "valid all identifier",
			identifier: &Identifier{
				Type:  AllIdentifier,
				Value: ".*",
			},
			wantError: false,
		},
		{
			name: "valid header identifier",
			identifier: &Identifier{
				Type:  Header,
				Value: "api-key",
			},
			wantError: false,
		},
		{
			name: "empty value",
			identifier: &Identifier{
				Type:  Header,
				Value: "",
			},
			wantError: true,
		},
		{
			name: "invalid identifier type",
			identifier: &Identifier{
				Type:  IdentifierType(999),
				Value: "api-key",
			},
			wantError: true,
			errorMsg:  "unsupported identifier type",
		},
		{
			name: "invalid regex value",
			identifier: &Identifier{
				Type:  Header,
				Value: "[invalid",
			},
			wantError: true,
			errorMsg:  "identifier value is not a valid regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.identifier)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateIdentifier() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateIdentifier() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateIdentifier() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name      string
		token     *Token
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil token",
			token:     nil,
			wantError: true,
			errorMsg:  "token cannot be nil",
		},
		{
			name: "valid token",
			token: &Token{
				Number:        1000,
				CountStrategy: TotalTokens,
			},
			wantError: false,
		},
		{
			name: "zero token number",
			token: &Token{
				Number:        0,
				CountStrategy: TotalTokens,
			},
			wantError: false,
		},
		{
			name: "negative token number",
			token: &Token{
				Number:        -100,
				CountStrategy: TotalTokens,
			},
			wantError: true,
			errorMsg:  "token number must be positive",
		},
		{
			name: "maximum valid token number",
			token: &Token{
				Number:        1000000000, // exactly 1 billion
				CountStrategy: TotalTokens,
			},
			wantError: false,
		},
		{
			name: "valid input tokens strategy",
			token: &Token{
				Number:        1000,
				CountStrategy: InputTokens,
			},
			wantError: false,
		},
		{
			name: "valid output tokens strategy",
			token: &Token{
				Number:        1000,
				CountStrategy: OutputTokens,
			},
			wantError: false,
		},
		{
			name: "invalid count strategy",
			token: &Token{
				Number:        1000,
				CountStrategy: CountStrategy(999),
			},
			wantError: true,
			errorMsg:  "unsupported count strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToken(tt.token)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateToken() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateToken() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateToken() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateTime(t *testing.T) {
	tests := []struct {
		name      string
		time      *Time
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil time",
			time:      nil,
			wantError: true,
			errorMsg:  "time cannot be nil",
		},
		{
			name: "valid second time unit",
			time: &Time{
				Unit:  Second,
				Value: 30,
			},
			wantError: false,
		},
		{
			name: "valid minute time unit",
			time: &Time{
				Unit:  Minute,
				Value: 5,
			},
			wantError: false,
		},
		{
			name: "valid hour time unit",
			time: &Time{
				Unit:  Hour,
				Value: 2,
			},
			wantError: false,
		},
		{
			name: "valid day time unit",
			time: &Time{
				Unit:  Day,
				Value: 1,
			},
			wantError: false,
		},
		{
			name: "invalid time unit",
			time: &Time{
				Unit:  TimeUnit(999),
				Value: 1,
			},
			wantError: true,
			errorMsg:  "unsupported time unit",
		},
		{
			name: "zero time value",
			time: &Time{
				Unit:  Minute,
				Value: 0,
			},
			wantError: false,
		},
		{
			name: "negative time value",
			time: &Time{
				Unit:  Minute,
				Value: -5,
			},
			wantError: true,
			errorMsg:  "time value must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTime(tt.time)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateTime() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateTime() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateTime() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateKeyItem(t *testing.T) {
	tests := []struct {
		name      string
		keyItem   *KeyItem
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil key item",
			keyItem:   nil,
			wantError: true,
			errorMsg:  "keyItem cannot be nil",
		},
		{
			name: "valid key item",
			keyItem: &KeyItem{
				Key: "user-*",
				Token: Token{
					Number:        1000,
					CountStrategy: TotalTokens,
				},
				Time: Time{
					Unit:  Minute,
					Value: 1,
				},
			},
			wantError: false,
		},
		{
			name: "empty key",
			keyItem: &KeyItem{
				Key: "",
				Token: Token{
					Number:        1000,
					CountStrategy: TotalTokens,
				},
				Time: Time{
					Unit:  Minute,
					Value: 1,
				},
			},
			wantError: true,
		},
		{
			name: "invalid key regex",
			keyItem: &KeyItem{
				Key: "[invalid",
				Token: Token{
					Number:        1000,
					CountStrategy: TotalTokens,
				},
				Time: Time{
					Unit:  Minute,
					Value: 1,
				},
			},
			wantError: true,
			errorMsg:  "key pattern is not a valid regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKeyItem(tt.keyItem)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateKeyItem() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateKeyItem() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateKeyItem() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateSpecificItem(t *testing.T) {
	tests := []struct {
		name         string
		specificItem *SpecificItem
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "nil rule item",
			specificItem: nil,
			wantError:    true,
			errorMsg:     "specificItem cannot be nil",
		},
		{
			name: "valid rule item",
			specificItem: &SpecificItem{
				Identifier: Identifier{
					Type:  Header,
					Value: "api-key",
				},
				KeyItems: []*KeyItem{
					{
						Key: "user-*",
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
			wantError: false,
		},
		{
			name: "empty key items",
			specificItem: &SpecificItem{
				Identifier: Identifier{
					Type:  Header,
					Value: "api-key",
				},
				KeyItems: []*KeyItem{},
			},
			wantError: true,
			errorMsg:  "keyItems cannot be empty",
		},
		{
			name: "invalid identifier",
			specificItem: &SpecificItem{
				Identifier: Identifier{
					Type:  IdentifierType(999),
					Value: "api-key",
				},
				KeyItems: []*KeyItem{
					{
						Key: "user-*",
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
			wantError: true,
			errorMsg:  "invalid identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpecificItem(tt.specificItem)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateSpecificItem() expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateSpecificItem() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSpecificItem() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		func() bool {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
