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
	"sync"
	"testing"
)

// Test helper functions to create test data
func createTestRuleForRuleManager(resource string) *Rule {
	return &Rule{
		ID:       "test-rule-1",
		Resource: resource,
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

func createTestRuleWithID(resource, id string) *Rule {
	rule := createTestRuleForRuleManager(resource)
	rule.ID = id
	return rule
}

func resetGlobalState() {
	ruleMap = make(map[string][]*Rule)
	currentRules = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.Mutex)
	rwMux = &sync.RWMutex{}
}

// Test LoadRules function
func TestLoadRules(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		rules         []*Rule
		expectUpdated bool
		expectError   bool
		expectedRules int
		setupExisting bool
		existingRules []*Rule
	}{
		{
			name: "Load valid rules",
			rules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
				createTestRuleForRuleManager("test-resource-2"),
			},
			expectUpdated: true,
			expectError:   false,
			expectedRules: 2,
		},
		{
			name:          "Load empty rules",
			rules:         []*Rule{},
			setupExisting: true,
			existingRules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
			},
			expectUpdated: true,
			expectError:   false,
			expectedRules: 0,
		},
		{
			name:          "Load nil rules",
			rules:         nil,
			setupExisting: true,
			existingRules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
			},
			expectUpdated: true,
			expectError:   false,
			expectedRules: 0,
		},
		{
			name: "Load same rules twice (should not update)",
			rules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
			},
			setupExisting: true,
			existingRules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
			},
			expectUpdated: false,
			expectError:   false,
			expectedRules: 1,
		},
		{
			name: "Load invalid rules (should be filtered out)",
			rules: []*Rule{
				createTestRuleForRuleManager("test-resource-1"),
				nil, // Invalid rule
				{
					Resource: "", // Invalid resource
					Strategy: PETA,
				},
			},
			expectUpdated: true,
			expectError:   false,
			expectedRules: 1,
		},
		{
			name: "Load rules with duplicate resources (last one wins)",
			rules: []*Rule{
				createTestRuleWithID("same-resource", "rule-1"),
				createTestRuleWithID("same-resource", "rule-2"),
			},
			expectUpdated: true,
			expectError:   false,
			expectedRules: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup existing rules if needed
			if tt.setupExisting {
				_, err := LoadRules(tt.existingRules)
				if err != nil {
					t.Fatalf("Failed to setup existing rules: %v", err)
				}
			}

			// Load test rules
			updated, err := LoadRules(tt.rules)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			// Check update expectation
			if updated != tt.expectUpdated {
				t.Errorf("Expected updated=%v, got %v", tt.expectUpdated, updated)
			}

			// Check rules count
			allRules := GetRules()
			if len(allRules) != tt.expectedRules {
				t.Errorf("Expected %d rules, got %d", tt.expectedRules, len(allRules))
			}
		})
	}
}

// Test onRuleUpdate function
func TestOnRuleUpdate(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name         string
		rawRulesMap  map[string][]*Rule
		expectError  bool
		expectedSize int
	}{
		{
			name: "Valid rules map",
			rawRulesMap: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1")},
				"resource-2": {createTestRuleForRuleManager("resource-2")},
			},
			expectError:  false,
			expectedSize: 2,
		},
		{
			name:         "Empty rules map",
			rawRulesMap:  map[string][]*Rule{},
			expectError:  false,
			expectedSize: 0,
		},
		{
			name: "Rules map with empty rule slices (should be filtered out)",
			rawRulesMap: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1")},
				"resource-2": {}, // Empty slice
			},
			expectError:  false,
			expectedSize: 1,
		},
		{
			name:         "Nil rules map",
			rawRulesMap:  nil,
			expectError:  false,
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			err := onRuleUpdate(tt.rawRulesMap)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			// Check rules map size
			rwMux.RLock()
			actualSize := len(ruleMap)
			rwMux.RUnlock()

			if actualSize != tt.expectedSize {
				t.Errorf("Expected ruleMap size %d, got %d", tt.expectedSize, actualSize)
			}

			// Verify currentRules is updated
			if !reflect.DeepEqual(currentRules, tt.rawRulesMap) {
				t.Error("currentRules was not updated correctly")
			}
		})
	}
}

// Test LoadRulesOfResource function
func TestLoadRulesOfResource(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		resource      string
		rules         []*Rule
		expectUpdated bool
		expectError   bool
		setupExisting bool
		existingRules []*Rule
	}{
		{
			name:     "Load valid rules for resource",
			resource: "test-resource",
			rules: []*Rule{
				createTestRuleForRuleManager("test-resource"),
			},
			expectUpdated: true,
			expectError:   false,
		},
		{
			name:          "Empty resource name",
			resource:      "",
			rules:         []*Rule{createTestRuleForRuleManager("test-resource")},
			expectUpdated: false,
			expectError:   true,
		},
		{
			name:          "Clear rules for resource",
			resource:      "test-resource",
			rules:         []*Rule{},
			expectUpdated: true,
			expectError:   false,
		},
		{
			name:          "Clear rules for resource with nil",
			resource:      "test-resource",
			rules:         nil,
			expectUpdated: true,
			expectError:   false,
		},
		{
			name:     "Load same rules for resource (should not update)",
			resource: "test-resource",
			rules: []*Rule{
				createTestRuleForRuleManager("test-resource"),
			},
			setupExisting: true,
			existingRules: []*Rule{
				createTestRuleForRuleManager("test-resource"),
			},
			expectUpdated: false,
			expectError:   false,
		},
		{
			name:     "Load different rules for resource (should update)",
			resource: "test-resource",
			rules: []*Rule{
				createTestRuleWithID("test-resource", "new-rule"),
			},
			setupExisting: true,
			existingRules: []*Rule{
				createTestRuleWithID("test-resource", "old-rule"),
			},
			expectUpdated: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup existing rules if needed
			if tt.setupExisting {
				_, err := LoadRulesOfResource(tt.resource, tt.existingRules)
				if err != nil {
					t.Fatalf("Failed to setup existing rules: %v", err)
				}
			}

			// Load test rules
			updated, err := LoadRulesOfResource(tt.resource, tt.rules)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			// Check update expectation
			if updated != tt.expectUpdated {
				t.Errorf("Expected updated=%v, got %v", tt.expectUpdated, updated)
			}

			// If no error and rules were provided, verify they were loaded
			if !tt.expectError && len(tt.rules) > 0 {
				resourceRules := GetRulesOfResource(tt.resource)
				if len(resourceRules) == 0 {
					t.Error("Expected rules to be loaded for resource, got none")
				}
			}
		})
	}
}

// Test onResourceRuleUpdate function
func TestOnResourceRuleUpdate(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name         string
		resource     string
		rawResRules  []*Rule
		expectError  bool
		expectExists bool
	}{
		{
			name:     "Update resource with valid rules",
			resource: "test-resource",
			rawResRules: []*Rule{
				createTestRuleForRuleManager("test-resource"),
			},
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "Update resource with empty rules (should delete)",
			resource:     "test-resource",
			rawResRules:  []*Rule{},
			expectError:  false,
			expectExists: false,
		},
		{
			name:         "Update resource with nil rules (should delete)",
			resource:     "test-resource",
			rawResRules:  nil,
			expectError:  false,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			err := onResourceRuleUpdate(tt.resource, tt.rawResRules)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			// Check if resource exists in ruleMap
			rwMux.RLock()
			_, exists := ruleMap[tt.resource]
			rwMux.RUnlock()

			if exists != tt.expectExists {
				t.Errorf("Expected resource exists=%v, got %v", tt.expectExists, exists)
			}

			// Check currentRules
			currentResourceRules, exists := currentRules[tt.resource]
			if tt.expectExists {
				if !exists {
					t.Error("Expected resource to exist in currentRules")
				}
				if !reflect.DeepEqual(currentResourceRules, tt.rawResRules) {
					t.Error("currentRules was not updated correctly")
				}
			}
		})
	}
}

// Test ClearRules function
func TestClearRules(t *testing.T) {
	defer resetGlobalState()

	// Setup some rules first
	rules := []*Rule{
		createTestRuleForRuleManager("resource-1"),
		createTestRuleForRuleManager("resource-2"),
	}
	_, err := LoadRules(rules)
	if err != nil {
		t.Fatalf("Failed to setup test rules: %v", err)
	}

	// Verify rules exist
	if len(GetRules()) == 0 {
		t.Fatal("Expected rules to be loaded, got none")
	}

	// Clear rules
	err = ClearRules()
	if err != nil {
		t.Errorf("Expected no error clearing rules, got %v", err)
	}

	// Verify rules are cleared
	if len(GetRules()) != 0 {
		t.Error("Expected all rules to be cleared, but some remain")
	}
}

// Test ClearRulesOfResource function
func TestClearRulesOfResource(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		resource      string
		expectError   bool
		setupRules    []*Rule
		otherResource string
	}{
		{
			name:     "Clear existing resource",
			resource: "test-resource",
			setupRules: []*Rule{
				createTestRuleForRuleManager("test-resource"),
				createTestRuleForRuleManager("other-resource"),
			},
			otherResource: "other-resource",
			expectError:   false,
		},
		{
			name:        "Clear non-existing resource",
			resource:    "non-existing",
			setupRules:  []*Rule{createTestRuleForRuleManager("test-resource")},
			expectError: false,
		},
		{
			name:        "Clear with empty resource name",
			resource:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup rules if provided
			if len(tt.setupRules) > 0 {
				_, err := LoadRules(tt.setupRules)
				if err != nil {
					t.Fatalf("Failed to setup test rules: %v", err)
				}
			}

			// Clear specific resource
			err := ClearRulesOfResource(tt.resource)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}

			// Verify target resource is cleared
			resourceRules := GetRulesOfResource(tt.resource)
			if len(resourceRules) != 0 {
				t.Errorf("Expected resource rules to be cleared, got %d", len(resourceRules))
			}

			// Verify other resources remain if they exist
			if tt.otherResource != "" {
				otherRules := GetRulesOfResource(tt.otherResource)
				if len(otherRules) == 0 {
					t.Error("Expected other resource rules to remain, but they were cleared")
				}
			}
		})
	}
}

// Test GetRules function
func TestGetRules(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		setupRules    []*Rule
		expectedCount int
	}{
		{
			name:          "Get rules when none exist",
			setupRules:    nil,
			expectedCount: 0,
		},
		{
			name: "Get rules when some exist",
			setupRules: []*Rule{
				createTestRuleForRuleManager("resource-1"),
				createTestRuleForRuleManager("resource-2"),
			},
			expectedCount: 2,
		},
		{
			name: "Get rules with duplicate resources (should return unique)",
			setupRules: []*Rule{
				createTestRuleWithID("same-resource", "rule-1"),
				createTestRuleWithID("same-resource", "rule-2"),
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup rules if provided
			if len(tt.setupRules) > 0 {
				_, err := LoadRules(tt.setupRules)
				if err != nil {
					t.Fatalf("Failed to setup test rules: %v", err)
				}
			}

			// Get rules
			rules := GetRules()

			// Check count
			if len(rules) != tt.expectedCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
			}

			// Verify returned rules are copies (not pointers)
			for _, rule := range rules {
				// This should not be a pointer dereference but a struct
				if reflect.TypeOf(rule).Kind() == reflect.Ptr {
					t.Error("Expected rule structs, got pointers")
				}
			}
		})
	}
}

// Test GetRulesOfResource function
func TestGetRulesOfResource(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		setupRules    []*Rule
		queryResource string
		expectedCount int
	}{
		{
			name:          "Get rules for non-existing resource",
			setupRules:    []*Rule{createTestRuleForRuleManager("other-resource")},
			queryResource: "test-resource",
			expectedCount: 0,
		},
		{
			name:          "Get rules for existing resource",
			setupRules:    []*Rule{createTestRuleForRuleManager("test-resource")},
			queryResource: "test-resource",
			expectedCount: 1,
		},
		{
			name: "Get rules with pattern matching",
			setupRules: []*Rule{
				createTestRuleForRuleManager("test-.*"),
				createTestRuleForRuleManager("other-resource"),
			},
			queryResource: "test-service",
			expectedCount: 1,
		},
		{
			name: "Get rules for multiple matching patterns",
			setupRules: []*Rule{
				createTestRuleForRuleManager("test-.*"),
				createTestRuleForRuleManager(".*-service"),
			},
			queryResource: "test-service",
			expectedCount: 2,
		},
		{
			name:          "Get rules when no rules exist",
			setupRules:    nil,
			queryResource: "any-resource",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup rules if provided
			if len(tt.setupRules) > 0 {
				_, err := LoadRules(tt.setupRules)
				if err != nil {
					t.Fatalf("Failed to setup test rules: %v", err)
				}
			}

			// Get rules for resource
			rules := GetRulesOfResource(tt.queryResource)

			// Check count
			if len(rules) != tt.expectedCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
			}

			// Verify returned rules are copies (not pointers)
			for _, rule := range rules {
				if reflect.TypeOf(rule).Kind() == reflect.Ptr {
					t.Error("Expected rule structs, got pointers")
				}
			}
		})
	}
}

// Test getRules function
func TestGetRulesInternal(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		setupRules    map[string][]*Rule
		expectedCount int
	}{
		{
			name:          "Get rules when ruleMap is empty",
			setupRules:    map[string][]*Rule{},
			expectedCount: 0,
		},
		{
			name: "Get rules when ruleMap has entries",
			setupRules: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1")},
				"resource-2": {createTestRuleForRuleManager("resource-2")},
			},
			expectedCount: 2,
		},
		{
			name: "Get rules with nil entries (should be filtered)",
			setupRules: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1"), nil},
				"resource-2": {nil, createTestRuleForRuleManager("resource-2")},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup ruleMap directly
			rwMux.Lock()
			ruleMap = tt.setupRules
			rwMux.Unlock()

			// Get rules
			rules := getRules()

			// Check count
			if len(rules) != tt.expectedCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
			}

			// Verify all returned rules are non-nil
			for _, rule := range rules {
				if rule == nil {
					t.Error("Found nil rule in results")
				}
			}
		})
	}
}

// Test getRulesOfResource function
func TestGetRulesOfResourceInternal(t *testing.T) {
	defer resetGlobalState()

	tests := []struct {
		name          string
		setupRules    map[string][]*Rule
		queryResource string
		expectedCount int
	}{
		{
			name:          "Get rules from empty ruleMap",
			setupRules:    map[string][]*Rule{},
			queryResource: "test-resource",
			expectedCount: 0,
		},
		{
			name: "Get rules with exact match",
			setupRules: map[string][]*Rule{
				"test-resource": {createTestRuleForRuleManager("test-resource")},
				"other":         {createTestRuleForRuleManager("other")},
			},
			queryResource: "test-resource",
			expectedCount: 1,
		},
		{
			name: "Get rules with regex pattern match",
			setupRules: map[string][]*Rule{
				"test-.*": {createTestRuleForRuleManager("test-.*")},
				"other":   {createTestRuleForRuleManager("other")},
			},
			queryResource: "test-service",
			expectedCount: 1,
		},
		{
			name: "Get rules with nil entries (should be filtered)",
			setupRules: map[string][]*Rule{
				"test-resource": {createTestRuleForRuleManager("test-resource"), nil},
			},
			queryResource: "test-resource",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalState()

			// Setup ruleMap directly
			rwMux.Lock()
			ruleMap = tt.setupRules
			rwMux.Unlock()

			// Get rules for resource
			rules := getRulesOfResource(tt.queryResource)

			// Check count
			if len(rules) != tt.expectedCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
			}

			// Verify all returned rules are non-nil
			for _, rule := range rules {
				if rule == nil {
					t.Error("Found nil rule in results")
				}
			}
		})
	}
}

// Test rulesFrom function
func TestRulesFrom(t *testing.T) {
	tests := []struct {
		name          string
		rulesMap      map[string][]*Rule
		expectedCount int
	}{
		{
			name:          "Empty map",
			rulesMap:      map[string][]*Rule{},
			expectedCount: 0,
		},
		{
			name:          "Nil map",
			rulesMap:      nil,
			expectedCount: 0,
		},
		{
			name: "Map with valid rules",
			rulesMap: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1")},
				"resource-2": {createTestRuleForRuleManager("resource-2")},
			},
			expectedCount: 2,
		},
		{
			name: "Map with nil rules (should be filtered)",
			rulesMap: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1"), nil},
				"resource-2": {nil, createTestRuleForRuleManager("resource-2")},
			},
			expectedCount: 2,
		},
		{
			name: "Map with empty rule slices",
			rulesMap: map[string][]*Rule{
				"resource-1": {},
				"resource-2": {createTestRuleForRuleManager("resource-2")},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := rulesFrom(tt.rulesMap)

			// Check count
			if len(rules) != tt.expectedCount {
				t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
			}

			// Verify all returned rules are non-nil
			for _, rule := range rules {
				if rule == nil {
					t.Error("Found nil rule in results")
				}
			}
		})
	}
}

// Test logRuleUpdate function
func TestLogRuleUpdate(t *testing.T) {
	tests := []struct {
		name     string
		rulesMap map[string][]*Rule
		verify   func() bool // Function to verify log output
	}{
		{
			name:     "Log with empty rules (should log cleared message)",
			rulesMap: map[string][]*Rule{},
			verify: func() bool {
				// In a real implementation, you might want to capture log output
				// For now, we just verify the function doesn't panic
				return true
			},
		},
		{
			name: "Log with rules (should log loaded message)",
			rulesMap: map[string][]*Rule{
				"resource-1": {createTestRuleForRuleManager("resource-1")},
			},
			verify: func() bool {
				return true
			},
		},
		{
			name:     "Log with nil map",
			rulesMap: nil,
			verify: func() bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function primarily logs, so we test it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("logRuleUpdate panicked: %v", r)
				}
			}()

			logRuleUpdate(tt.rulesMap)

			if !tt.verify() {
				t.Error("Log verification failed")
			}
		})
	}
}

// Test concurrent access to rule manager
func TestRuleManager_ConcurrentAccess(t *testing.T) {
	defer resetGlobalState()

	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 types of operations

	// Concurrent rule loading
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				rule := createTestRuleWithID("concurrent-resource",
					strings.Join([]string{"rule", string(rune(id)), string(rune(j))}, "-"))
				LoadRules([]*Rule{rule})
			}
		}(i)
	}

	// Concurrent rule reading
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				GetRules()
				GetRulesOfResource("concurrent-resource")
			}
		}()
	}

	// Concurrent resource-specific operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				resource := strings.Join([]string{"resource", string(rune(id))}, "-")
				rule := createTestRuleForRuleManager(resource)
				LoadRulesOfResource(resource, []*Rule{rule})
			}
		}(i)
	}

	wg.Wait()
}

// Benchmark tests
func BenchmarkLoadRules(b *testing.B) {
	defer resetGlobalState()

	rules := []*Rule{
		createTestRuleForRuleManager("benchmark-resource-1"),
		createTestRuleForRuleManager("benchmark-resource-2"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadRules(rules)
	}
}

func BenchmarkGetRules(b *testing.B) {
	defer resetGlobalState()

	// Setup some rules
	rules := []*Rule{
		createTestRuleForRuleManager("benchmark-resource-1"),
		createTestRuleForRuleManager("benchmark-resource-2"),
		createTestRuleForRuleManager("benchmark-resource-3"),
	}
	LoadRules(rules)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRules()
	}
}

func BenchmarkGetRulesOfResource(b *testing.B) {
	defer resetGlobalState()

	// Setup some rules
	rules := []*Rule{
		createTestRuleForRuleManager("benchmark-.*"),
		createTestRuleForRuleManager("other-resource"),
	}
	LoadRules(rules)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRulesOfResource("benchmark-service")
	}
}

func BenchmarkRulesFrom(b *testing.B) {
	rulesMap := map[string][]*Rule{
		"resource-1": {createTestRuleForRuleManager("resource-1")},
		"resource-2": {createTestRuleForRuleManager("resource-2")},
		"resource-3": {createTestRuleForRuleManager("resource-3")},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rulesFrom(rulesMap)
	}
}
