package base

import "regexp"

type StringMatch struct {
	Exact  string
	Prefix string
	Regex  string
}

func (s *StringMatch) IsMatch(input string) bool {
	if input == "" {
		return false
	}

	if s.Exact != "" {
		return input == s.Exact
	} else if s.Prefix != "" {
		return len(input) >= len(s.Prefix) && input[:len(s.Prefix)] == s.Prefix
	} else if s.Regex != "" {
		matched, _ := regexp.MatchString(s.Regex, input)
		return matched
	}
	return true
}
