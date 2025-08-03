package base

import (
	"net/url"
	"strconv"
)

type HTTPMatchRequest struct {
	Name        string
	Headers     map[string]*StringMatch
	Uri         *StringMatch
	Scheme      *StringMatch
	Authority   *StringMatch
	Method      *StringMatch
	Port        *int
	QueryParams map[string]*StringMatch
}

// TODO
func (h *HTTPMatchRequest) IsMatch(context *TrafficContext) bool {

	for key, match := range h.Headers {
		if v, ok := context.Headers[key]; ok && !match.IsMatch(v) {
			return false
		}
	}

	if h.Uri != nil && !h.Uri.IsMatch(context.Uri) {
		return false
	}

	var parsedURL *url.URL
	var err error

	if h.Uri != nil || h.Scheme != nil || h.Authority != nil || h.Port != nil {
		parsedURL, err = url.Parse(context.Uri)
		if err != nil {
			return false
		}
	}
	if h.Uri != nil && !h.Uri.IsMatch(parsedURL.Path) {
		return false
	}
	if h.Scheme != nil && !h.Scheme.IsMatch(parsedURL.Scheme) {
		return false
	}
	if h.Authority != nil && !h.Authority.IsMatch(parsedURL.Host) {
		return false
	}
	if h.Port != nil {
		p, err := strconv.Atoi(parsedURL.Port())
		if err != nil || *h.Port != p {
			return false
		}
	}
	if !h.Method.IsMatch(context.MethodName) {
		return false
	}

	return true
}
