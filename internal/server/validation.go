package server

import (
	"regexp"

	"github.com/go-playground/validator"
)

// validUsernameRegex is the DNS-1123 subdomain regex: usernames are used as ServiceAccount names.
// The dots are escaped on purpose: an unescaped "." would accept any character (see PR #182).
const validUsernameRegex = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"

// maxUsernameLength is the maximum length of a DNS-1123 subdomain (and of a ServiceAccount name).
const maxUsernameLength = 253

var invalidUsernameError = `username must be DNS-1123 compliant, it must consist of lower case alphanumeric characters, '-' or '.', must start and end with an alphanumeric character and be at most 253 characters long (e.g. 'gino.mycompany'), regex used for validation is ` + validUsernameRegex

var usernameRe = regexp.MustCompile(validUsernameRegex)

func isValidUsername(username string) (valid bool) {
	return len(username) <= maxUsernameLength && usernameRe.MatchString(username)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)

}
