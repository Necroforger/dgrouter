package dgrouter

import "regexp"

// NewRegexMatcher returns a new regex matcher
func NewRegexMatcher(regex string) func(string) bool {
	r := regexp.MustCompile(regex)
	return func(command string) bool {
		return r.MatchString(command)
	}
}

// NewNameMatcher returns a matcher that matches a route's name
func NewNameMatcher(r *Route) func(string) bool {
	return func(command string) bool {
		return command == r.Name
	}
}
