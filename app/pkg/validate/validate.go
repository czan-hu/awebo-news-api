// Package validate holds small format checks that are simpler to express as
// plain regexes than as custom go-playground/validator tags.
package validate

import "regexp"

var (
	usernameRe = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
	codeRe     = regexp.MustCompile(`^[0-9]{6}$`)
)

func Username(s string) bool {
	return usernameRe.MatchString(s)
}

func Code(s string) bool {
	return codeRe.MatchString(s)
}
