package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateConfirmEmailInput(t *testing.T) {
	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		code      string
	}{
		{
			name:      "empty code",
			expectErr: "Code is required",
			argName:   "code",
			code:      "",
		},
		{
			name:  "valid code",
			valid: true,
			code:  "some-valid-code",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := service.ValidateConfirmEmailInput(&c.code)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
