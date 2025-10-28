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
	"regexp"
)

func IsValidRule(r *Rule) error {
	if r == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	// Validate Resource
	if err := validateResource(r.Resource); err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}

	// Validate Strategy
	if err := validateStrategy(r.Strategy); err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}

	// Validate Encoding
	if err := validateEncoding(r.Encoding); err != nil {
		return fmt.Errorf("invalid token encoding: %w", err)
	}

	// Validate SpecificItems (required)
	if len(r.SpecificItems) == 0 {
		return fmt.Errorf("specificItems cannot be empty")
	}

	for i, specificItem := range r.SpecificItems {
		if err := validateSpecificItem(specificItem); err != nil {
			return fmt.Errorf("invalid specificItem[%d]: %w", i, err)
		}
	}

	return nil
}

func validateResource(resource string) error {
	if resource == "" {
		return fmt.Errorf("resource pattern cannot be empty")
	}

	if _, err := regexp.Compile(resource); err != nil {
		return fmt.Errorf("resource pattern is not a valid regex: %w", err)
	}

	return nil
}

func validateStrategy(strategy Strategy) error {
	switch strategy {
	case FixedWindow, PETA:
		return nil
	default:
		return fmt.Errorf("unsupported strategy: %s", strategy.String())
	}
}

func validateEncoding(encoding TokenEncoding) error {
	// Validate TokenEncoding
	switch encoding.Provider {
	case OpenAIEncoderProvider:
		return nil
	default:
		return fmt.Errorf("unsupported token encoding provider: %v", encoding.Provider.String())
	}
}

func validateSpecificItem(specificItem *SpecificItem) error {
	if specificItem == nil {
		return fmt.Errorf("specificItem cannot be nil")
	}

	// Validate Identifier
	if err := validateIdentifier(&specificItem.Identifier); err != nil {
		return fmt.Errorf("invalid identifier: %w", err)
	}

	// Validate KeyItems (required)
	if len(specificItem.KeyItems) == 0 {
		return fmt.Errorf("keyItems cannot be empty")
	}

	for i, keyItem := range specificItem.KeyItems {
		if err := validateKeyItem(keyItem); err != nil {
			return fmt.Errorf("invalid keyItem[%d]: %w", i, err)
		}
	}

	return nil
}

func validateIdentifier(identifier *Identifier) error {
	if identifier == nil {
		return fmt.Errorf("identifier cannot be nil")
	}

	// Validate Type
	switch identifier.Type {
	case AllIdentifier, Header:
		// Valid types
	default:
		return fmt.Errorf("unsupported identifier type: %v", identifier.Type)
	}

	if identifier.Value == "" {
		return fmt.Errorf("identifier value pattern cannot be empty")
	}
	if _, err := regexp.Compile(identifier.Value); err != nil {
		return fmt.Errorf("identifier value is not a valid regex: %w", err)
	}

	return nil
}

func validateKeyItem(keyItem *KeyItem) error {
	if keyItem == nil {
		return fmt.Errorf("keyItem cannot be nil")
	}

	if keyItem.Key == "" {
		return fmt.Errorf("key pattern cannot be empty")
	}
	if _, err := regexp.Compile(keyItem.Key); err != nil {
		return fmt.Errorf("key pattern is not a valid regex: %w", err)
	}

	// Validate Token (required)
	if err := validateToken(&keyItem.Token); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Validate Time (required)
	if err := validateTime(&keyItem.Time); err != nil {
		return fmt.Errorf("invalid time: %w", err)
	}

	return nil
}

func validateToken(token *Token) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	// Validate Number (required, must be positive)
	if token.Number < 0 {
		return fmt.Errorf("token number must be positive, got: %d", token.Number)
	}

	// Validate CountStrategy
	switch token.CountStrategy {
	case TotalTokens, InputTokens, OutputTokens:
		// Valid strategies
	default:
		return fmt.Errorf("unsupported count strategy: %v", token.CountStrategy)
	}

	return nil
}

func validateTime(time *Time) error {
	if time == nil {
		return fmt.Errorf("time cannot be nil")
	}

	// Validate Unit (required)
	switch time.Unit {
	case Second, Minute, Hour, Day:
		// Valid units
	default:
		return fmt.Errorf("unsupported time unit: %v", time.Unit)
	}

	// Validate Value (required, must be positive)
	if time.Value < 0 {
		return fmt.Errorf("time value must be positive, got: %d", time.Value)
	}

	return nil
}
