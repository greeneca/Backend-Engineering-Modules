package test_utils

import (
	"regexp"
	"fmt"
)

// RegexMatcher implements gomock.Matcher for regular expression matching.
type RegexMatcher struct {
	pattern *regexp.Regexp
}

// Matches returns true if the input string matches the regex pattern.
func (m RegexMatcher) Matches(in any) bool {
	s, ok := in.(string)
	if !ok {
		return false
	}
	return m.pattern.MatchString(s)
}

// String returns a description of the matcher.
func (m RegexMatcher) String() string {
	return fmt.Sprintf("matches regex /%s/", m.pattern.String())
}

// NewRegexMatcher creates a new RegexMatcher.
func NewRegexMatcher(pattern string) *RegexMatcher {
	return &RegexMatcher{
		pattern: regexp.MustCompile(pattern),
	}
}
