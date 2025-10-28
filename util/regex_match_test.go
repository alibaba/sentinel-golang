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

package util

import (
	"regexp"
	"testing"
)

func TestRegexMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
		want    bool
	}{
		{"empty pattern", "", "test", false},
		{"empty text", "test", "", false},
		{"both empty", "", "", true},
		{"asterisk pattern empty text", ".*", "", false},

		{"asterisk pattern", ".*", "anything", true},
		{"asterisk pattern special chars", ".*", "!@#$%^&*()", true},

		{"exact match", "hello", "hello", true},
		{"no match", "hello", "world", false},
		{"case sensitive", "Hello", "hello", false},
		{"partial match should fail", "hello", "hello world", false},
		{"partial match should fail 2", "world", "hello world", false},

		{"dot wildcard", "h.llo", "hello", true},
		{"dot wildcard", "h.llo", "hallo", true},
		{"dot wildcard no match", "h.llo", "hllo", false},
		{"dot wildcard partial should fail", "h.llo", "say hello", false},
		{"star quantifier", "hel*o", "helo", true},
		{"star quantifier", "hel*o", "hello", true},
		{"star quantifier multiple", "hel*o", "helllo", true},
		{"star quantifier partial should fail", "hel*o", "say hello", false},
		{"plus quantifier", "hel+o", "hello", true},
		{"plus quantifier no match", "hel+o", "heo", false},
		{"plus quantifier partial should fail", "hel+o", "say hello", false},

		{"digit class", "[0-9]+", "12345", true},
		{"digit class no match", "[0-9]+", "abc", false},
		{"digit class partial should fail", "[0-9]+", "abc123", false},
		{"letter class", "[a-z]+", "hello", true},
		{"letter class no match", "[a-z]+", "HELLO", false},
		{"letter class partial should fail", "[a-z]+", "hello123", false},
		{"mixed class", "[a-zA-Z0-9]+", "Hello123", true},
		{"mixed class partial should fail", "[a-zA-Z0-9]+", "Hello123!", false},

		{"start anchor exact", "^hello$", "hello", true},
		{"start anchor partial should fail", "^hello$", "hello world", false},
		{"start anchor only", "^hello", "hello", true},
		{"start anchor only partial should fail", "^hello", "hello world", false},
		{"end anchor only", "world$", "world", true},
		{"end anchor only partial should fail", "world$", "hello world", false},

		{"email pattern valid", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "test@example.com", true},
		{"email pattern invalid", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "invalid-email", false},
		{"email pattern without anchors", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, "test@example.com", true},
		{"phone pattern valid", `^\+?[1-9]\d{1,14}$`, "+1234567890", true},
		{"phone pattern invalid", `^\+?[1-9]\d{1,14}$`, "abc123", false},

		{"group match", "(hello|world)", "hello", true},
		{"group match alt", "(hello|world)", "world", true},
		{"group no match", "(hello|world)", "test", false},
		{"group partial should fail", "(hello|world)", "say hello", false},

		{"escaped dot", `hello\.world`, "hello.world", true},
		{"escaped dot no match", `hello\.world`, "helloXworld", false},
		{"escaped dot partial should fail", `hello\.world`, "say hello.world", false},
		{"escaped plus", `test\+`, "test+", true},
		{"escaped star", `test\*`, "test*", true},

		{"invalid pattern unclosed paren", "(hello", "hello", false},
		{"invalid pattern unclosed bracket", "[hello", "hello", false},
		{"invalid pattern bad escape", `\`, "test", false},
		{"invalid pattern bad quantifier", "*hello", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RegexMatch(tt.pattern, tt.text)
			if got != tt.want {
				t.Errorf("RegexMatch(%q, %q) = %v; want %v", tt.pattern, tt.text, got, tt.want)
			}
		})
	}
}

func TestEnsureExactRegexMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{"no anchors", "hello", "^hello$"},
		{"both anchors", "^hello$", "^hello$"},
		{"start anchor only", "^hello", "^hello$"},
		{"end anchor only", "hello$", "^hello$"},
		{"complex pattern", "[a-z]+", "^[a-z]+$"},
		{"already anchored complex", "^[a-z]+$", "^[a-z]+$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureExactRegexMatch(tt.pattern)
			if got != tt.expected {
				t.Errorf("ensureExactRegexMatch(%q) = %q; want %q", tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestGetCompiledRegex(t *testing.T) {
	cacheMu.Lock()
	regexCache = make(map[string]*regexp.Regexp)
	cacheMu.Unlock()

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid pattern", "^hello$", false},
		{"valid complex pattern", `^[a-z]+$`, false},
		{"invalid pattern", "(hello", true},
		{"invalid pattern bracket", "[hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := getCompiledRegex(tt.pattern)

			if tt.wantErr {
				if err == nil {
					t.Errorf("getCompiledRegex(%q) expected error, got nil", tt.pattern)
				}
				if regex != nil {
					t.Errorf("getCompiledRegex(%q) expected nil regex when error, got %v", tt.pattern, regex)
				}
			} else {
				if err != nil {
					t.Errorf("getCompiledRegex(%q) unexpected error: %v", tt.pattern, err)
				}
				if regex == nil {
					t.Errorf("getCompiledRegex(%q) expected regex, got nil", tt.pattern)
				}
			}
		})
	}
}

func TestRegexCaching(t *testing.T) {
	cacheMu.Lock()
	regexCache = make(map[string]*regexp.Regexp)
	cacheMu.Unlock()

	pattern := "^test_pattern_for_caching$"

	regex1, err1 := getCompiledRegex(pattern)
	if err1 != nil {
		t.Fatalf("First call failed: %v", err1)
	}

	cacheMu.RLock()
	cachedRegex, exists := regexCache[pattern]
	cacheMu.RUnlock()

	if !exists {
		t.Error("Pattern should be cached after first call")
	}

	if cachedRegex != regex1 {
		t.Error("Cached regex should be the same instance as returned")
	}

	regex2, err2 := getCompiledRegex(pattern)
	if err2 != nil {
		t.Fatalf("Second call failed: %v", err2)
	}

	if regex1 != regex2 {
		t.Error("Second call should return the same cached instance")
	}
}

func TestRegexMatchConcurrency(t *testing.T) {
	pattern := "concurrent_test_[0-9]+"
	text := "concurrent_test_123"

	cacheMu.Lock()
	regexCache = make(map[string]*regexp.Regexp)
	cacheMu.Unlock()

	const numGoroutines = 100
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result := RegexMatch(pattern, text)
			results <- result
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if !result {
			t.Error("All concurrent calls should return true")
		}
	}

	cacheMu.RLock()
	if len(regexCache) != 1 {
		t.Errorf("Expected 1 cached regex, got %d", len(regexCache))
	}
	cacheMu.RUnlock()
}

func BenchmarkRegexMatch(b *testing.B) {
	// 清空缓存
	cacheMu.Lock()
	regexCache = make(map[string]*regexp.Regexp)
	cacheMu.Unlock()

	pattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	text := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegexMatch(pattern, text)
	}
}

func BenchmarkRegexMatchCached(b *testing.B) {
	pattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	text := "test@example.com"

	RegexMatch(pattern, text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegexMatch(pattern, text)
	}
}

func BenchmarkRegexMatchWithoutCache(b *testing.B) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	text := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			b.Fatal(err)
		}
		regex.MatchString(text)
	}
}

func TestRegexMatchUnicode(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
		want    bool
	}{
		{"chinese exact match", "你好", "你好", true},
		{"chinese no match - partial", "你好", "你好世界", false},
		{"chinese no match - different", "你好", "世界", false},

		{"emoji exact match", "😀", "😀", true},
		{"emoji no match - in sentence", "😀", "Hello 😀 World", false},
		{"emoji no match", "😀", "Hello World", false},

		{"mixed unicode exact", "[\\p{Han}]+", "中文测试", true},
		{"mixed unicode no match", "[\\p{Han}]+", "English test", false},
		{"mixed unicode partial should fail", "[\\p{Han}]+", "English 中文 test", false},

		{"chinese with punctuation", "你好！", "你好！", true},
		{"chinese with punctuation partial", "你好！", "你好！世界", false},
		{"japanese hiragana", "こんにちは", "こんにちは", true},
		{"japanese hiragana partial", "こんにちは", "こんにちは世界", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RegexMatch(tt.pattern, tt.text)
			if got != tt.want {
				t.Errorf("RegexMatch(%q, %q) = %v; want %v", tt.pattern, tt.text, got, tt.want)
			}
		})
	}
}

func TestRegexMatchSpecialCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
		want    bool
	}{
		{"whitespace exact", "\\s+", "   ", true},
		{"whitespace partial should fail", "\\s+", "hello   world", false},

		{"number exact", "\\d+", "12345", true},
		{"number partial should fail", "\\d+", "abc12345xyz", false},

		{"word boundary", "\\bhello\\b", "hello", true},
		{"word boundary partial should fail", "\\bhello\\b", "say hello world", false},

		{"line start and end", "^test$", "test", true},
		{"line start and end multiline should fail", "^test$", "test\nmore", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RegexMatch(tt.pattern, tt.text)
			if got != tt.want {
				t.Errorf("RegexMatch(%q, %q) = %v; want %v", tt.pattern, tt.text, got, tt.want)
			}
		})
	}
}
