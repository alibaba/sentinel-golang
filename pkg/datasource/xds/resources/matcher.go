package resources

import (
	"regexp"
	"strings"

	v3 "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/config/route/v3"
	v3matcher "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/type/matcher/v3"
)

type PrefixMatcher string

func (pm PrefixMatcher) Match(other string) bool {
	return strings.HasPrefix(other, string(pm))
}

type ExactMatcher string

func (em ExactMatcher) Match(other string) bool {
	return string(em) == other
}

type RegexMatcher struct {
	re *regexp.Regexp
}

func (rm *RegexMatcher) Match(other string) bool {
	return rm.re.MatchString(other)
}

type Matcher interface {
	Match(string) bool
}

type Matchers map[string]Matcher

func (ms Matchers) Match(other map[string]string) bool {
	for key, m := range ms {
		if val, ok := other[key]; !ok || !m.Match(val) {
			return false
		}
	}
	return true
}

// BuildMatchers build matcher set from headers
func BuildMatchers(headers []*v3.HeaderMatcher) Matchers {
	ms := map[string]Matcher{}
	for _, header := range headers {
		switch hm := header.GetHeaderMatchSpecifier().(type) {
		case *v3.HeaderMatcher_StringMatch:
			switch p := hm.StringMatch.GetMatchPattern().(type) {
			case *v3matcher.StringMatcher_Exact:
				if p.Exact != "" {
					ms[header.Name] = ExactMatcher(p.Exact)
				}
			case *v3matcher.StringMatcher_Prefix:
				if p.Prefix != "" {
					ms[header.Name] = PrefixMatcher(p.Prefix)
				}
			case *v3matcher.StringMatcher_SafeRegex:
				// only support google re2
				if p.SafeRegex != nil && p.SafeRegex.Regex != "" {
					re2, err := regexp.Compile(p.SafeRegex.Regex)
					if err != nil {
						continue
					}
					ms[header.Name] = &RegexMatcher{
						re: re2,
					}
				}
			}
		}
	}
	return ms
}
