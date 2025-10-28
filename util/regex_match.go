package util

import (
	"regexp"
	"strings"
	"sync"
)

var (
	regexCache = make(map[string]*regexp.Regexp)
	cacheMu    sync.RWMutex
)

const (
	RegexBeginPattern = "^"
	RegexEndPattern   = "$"
)

func RegexMatch(pattern, text string) bool {
	if pattern == "" && text == "" {
		return true
	}
	if pattern == "" || text == "" {
		return false
	}

	exactPattern := ensureExactRegexMatch(pattern)

	regex, err := getCompiledRegex(exactPattern)
	if err != nil {
		return false
	}

	return regex.MatchString(text)
}

func ensureExactRegexMatch(pattern string) string {
	if strings.HasPrefix(pattern, RegexBeginPattern) && strings.HasSuffix(pattern, RegexEndPattern) {
		return pattern
	}

	if strings.HasPrefix(pattern, RegexBeginPattern) && !strings.HasSuffix(pattern, RegexEndPattern) {
		return pattern + RegexEndPattern
	}

	if !strings.HasPrefix(pattern, RegexBeginPattern) && strings.HasSuffix(pattern, RegexEndPattern) {
		return RegexBeginPattern + pattern
	}

	return RegexBeginPattern + pattern + RegexEndPattern
}

func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	cacheMu.RLock()
	if regex, exists := regexCache[pattern]; exists {
		cacheMu.RUnlock()
		return regex, nil
	}
	cacheMu.RUnlock()

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	regexCache[pattern] = regex
	cacheMu.Unlock()

	return regex, nil
}
