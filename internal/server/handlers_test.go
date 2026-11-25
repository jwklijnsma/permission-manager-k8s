package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		isValid  bool
	}{
		{name: "simple", username: "gino", isValid: true},
		{name: "with digits", username: "gino1", isValid: true},
		{name: "with dash", username: "gino-pino", isValid: true},
		{name: "example from the error message", username: "gino.mycompany", isValid: true},
		{name: "upper case", username: "Gino", isValid: false},
		{name: "space", username: "Gino Mycompany", isValid: false},
		{name: "email", username: "gino@pino", isValid: false},
		// regression for PR #182: an unescaped dot would accept these
		{name: "any char in place of dot", username: "gino_pino", isValid: false},
		{name: "slash (path traversal)", username: "gino/../pino", isValid: false},
		{name: "leading dash", username: "-gino", isValid: false},
		{name: "trailing dot", username: "gino.", isValid: false},
		{name: "empty", username: "", isValid: false},
		{name: "max length", username: strings.Repeat("a", 253), isValid: true},
		{name: "too long", username: strings.Repeat("a", 254), isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, isValidUsername(tt.username))
		})
	}
}
