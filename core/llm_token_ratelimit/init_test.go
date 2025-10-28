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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test concurrent safety of SafeConfig
func TestSafeConfig_ConcurrentSetGet(t *testing.T) {
	config := &SafeConfig{}
	const numGoroutines = 100
	const numOperations = 1000

	// Create test configs
	configs := make([]*Config, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		configs[i] = &Config{
			Redis: &Redis{
				Addrs: []*RedisAddr{
					{Name: fmt.Sprintf("redis-%d", i), Port: int32(6379 + i)},
				},
			},
			ErrorCode:    int32(i),
			ErrorMessage: fmt.Sprintf("error-%d", i),
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				if err := config.SetConfig(configs[id]); err != nil {
					errors <- fmt.Errorf("SetConfig failed in goroutine %d: %v", id, err)
					return
				}
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cfg := config.GetConfig()
				// Basic validation - should not panic or return corrupted data
				if cfg != nil {
					if cfg.Redis != nil {
						_ = cfg.Redis.Addrs // Access should not cause data race
					}
				}
			}
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errors:
		t.Fatal(err)
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}

	// Check for any remaining errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent error: %v", err)
	}
}

func TestSafeConfig_NilHandling(t *testing.T) {
	var nilConfig *SafeConfig

	// Test nil SafeConfig
	result := nilConfig.GetConfig()
	assert.Nil(t, result)

	err := nilConfig.SetConfig(&Config{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "safe config is nil")

	// Test setting nil config
	config := &SafeConfig{}
	err = config.SetConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestSafeConfig_GetAfterSet(t *testing.T) {
	config := &SafeConfig{}

	// Initially nil
	assert.Nil(t, config.GetConfig())

	// Set and get
	testConfig := &Config{
		ErrorCode:    500,
		ErrorMessage: "test error",
	}

	err := config.SetConfig(testConfig)
	require.NoError(t, err)

	retrieved := config.GetConfig()
	require.NotNil(t, retrieved)
	assert.Equal(t, int32(500), retrieved.ErrorCode)
	assert.Equal(t, "test error", retrieved.ErrorMessage)
}

// Test concurrent safety of global operations
func TestInit_ConcurrentSafety(t *testing.T) {
	// Save original state
	originalConfig := globalConfig.GetConfig()
	defer func() {
		if originalConfig != nil {
			globalConfig.SetConfig(originalConfig)
		}
	}()

	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Create different configs for each goroutine
	configs := make([]*Config, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		configs[i] = &Config{
			Redis: &Redis{
				Addrs: []*RedisAddr{
					{Name: "localhost", Port: 6379},
				},
				DialTimeout:  5000,
				ReadTimeout:  5000,
				WriteTimeout: 5000,
				PoolTimeout:  5000,
				PoolSize:     10,
				MinIdleConns: 5,
				MaxRetries:   3,
			},
			ErrorCode:    int32(400 + i),
			ErrorMessage: fmt.Sprintf("concurrent error %d", i),
		}
	}

	// Concurrent Init calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Note: Init might fail due to Redis connection, but should not cause data races
			if err := Init(configs[id]); err != nil {
				// Redis connection errors are expected in test environment
				if !containsRedisError(err) {
					errors <- fmt.Errorf("unexpected Init error in goroutine %d: %v", id, err)
				}
			}

			// Test reading config after init
			cfg := globalConfig.GetConfig()
			if cfg != nil {
				// Basic validation
				_ = cfg.ErrorCode
				_ = cfg.ErrorMessage
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errors:
		t.Fatal(err)
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}

	// Check for any remaining errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent error: %v", err)
	}
}

// Test that concurrent access doesn't cause data corruption
func TestSafeConfig_DataIntegrity(t *testing.T) {
	config := &SafeConfig{}

	// Predefined configs with distinct data
	configs := []*Config{
		{
			ErrorCode:    100,
			ErrorMessage: "config-100",
		},
		{
			ErrorCode:    200,
			ErrorMessage: "config-200",
		},
		{
			ErrorCode:    300,
			ErrorMessage: "config-300",
		},
	}

	const numIterations = 1000
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Writer goroutines - each writes a specific config repeatedly
	for configIndex, cfg := range configs {
		wg.Add(1)
		go func(configIdx int, safeConfig *SafeConfig, config *Config) {
			defer wg.Done()
			for i := 0; i < numIterations; i++ {
				if err := safeConfig.SetConfig(config); err != nil {
					errors <- fmt.Errorf("SetConfig failed for config %d: %v", configIdx, err)
					return
				}
			}
		}(configIndex, config, cfg)
	}

	// Reader goroutines - validate data integrity
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				cfg := config.GetConfig()
				if cfg != nil {
					// Validate data consistency - ErrorCode should match ErrorMessage pattern
					expectedMessage := fmt.Sprintf("config-%d", cfg.ErrorCode)
					if cfg.ErrorMessage != expectedMessage {
						errors <- fmt.Errorf("data corruption detected by reader %d: ErrorCode=%d, ErrorMessage=%s",
							readerID, cfg.ErrorCode, cfg.ErrorMessage)
						return
					}
				}
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errors:
		t.Fatal(err)
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}

	// Check for any remaining errors
	close(errors)
	for err := range errors {
		t.Errorf("Data integrity error: %v", err)
	}
}

// Benchmark concurrent access
func BenchmarkSafeConfig_ConcurrentSetGet(b *testing.B) {
	config := &SafeConfig{}
	testConfig := &Config{
		ErrorCode:    500,
		ErrorMessage: "benchmark config",
	}

	// Pre-set a config
	config.SetConfig(testConfig)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of reads and writes (90% reads, 10% writes)
			if b.N%10 == 0 {
				config.SetConfig(testConfig)
			} else {
				config.GetConfig()
			}
		}
	})
}

func BenchmarkSafeConfig_ReadOnly(b *testing.B) {
	config := &SafeConfig{}
	testConfig := &Config{
		ErrorCode:    500,
		ErrorMessage: "benchmark config",
	}
	config.SetConfig(testConfig)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			config.GetConfig()
		}
	})
}

func BenchmarkSafeConfig_WriteOnly(b *testing.B) {
	config := &SafeConfig{}
	testConfig := &Config{
		ErrorCode:    500,
		ErrorMessage: "benchmark config",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.SetConfig(testConfig)
	}
}

// Test global config concurrency
func TestGlobalConfig_ConcurrentAccess(t *testing.T) {
	// Save original state
	originalConfig := globalConfig.GetConfig()
	defer func() {
		if originalConfig != nil {
			globalConfig.SetConfig(originalConfig)
		}
	}()

	const numGoroutines = 20
	const numOperations = 500

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Test config
	testConfig := &Config{
		ErrorCode:    999,
		ErrorMessage: "global test config",
	}

	// Concurrent access to global config
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Mix operations
				switch j % 3 {
				case 0:
					// Set config
					if err := globalConfig.SetConfig(testConfig); err != nil {
						errors <- fmt.Errorf("globalConfig.SetConfig failed in goroutine %d: %v", id, err)
						return
					}
				case 1:
					// Get config
					cfg := globalConfig.GetConfig()
					if cfg != nil {
						_ = cfg.ErrorCode // Basic access
					}
				}
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errors:
		t.Fatal(err)
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}

	// Check for any remaining errors
	close(errors)
	for err := range errors {
		t.Errorf("Global config concurrent error: %v", err)
	}
}

// Test edge cases with nil values
func TestSafeConfig_NilValuesConcurrency(t *testing.T) {

	const numGoroutines = 20
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Concurrent operations with nil handling
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			config := &SafeConfig{}
			// Try to set nil config (should fail gracefully)
			err := config.SetConfig(nil)
			if err == nil {
				errors <- fmt.Errorf("SetConfig should reject nil config in goroutine %d", id)
				return
			}

			// Get config when nothing is set
			cfg := config.GetConfig()
			if cfg != nil {
				errors <- fmt.Errorf("GetConfig should return nil when no config is set in goroutine %d", id)
				return
			}

			// Set valid config
			validConfig := &Config{ErrorCode: int32(id)}
			if err := config.SetConfig(validConfig); err != nil {
				errors <- fmt.Errorf("SetConfig failed for valid config in goroutine %d: %v", id, err)
				return
			}

			// Get config again
			cfg = config.GetConfig()
			if cfg == nil {
				errors <- fmt.Errorf("GetConfig returned nil after setting valid config in goroutine %d", id)
				return
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errors:
		t.Fatal(err)
	case <-time.After(15 * time.Second):
		t.Fatal("Test timed out")
	}

	// Check for any remaining errors
	close(errors)
	for err := range errors {
		t.Errorf("Nil values concurrency error: %v", err)
	}
}

// Helper function to check if error is related to Redis connection
func containsRedisError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsAny(errStr, []string{
		"failed to connect to redis",
		"redis client",
		"connection refused",
		"timeout",
		"no route to host",
	})
}

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}
