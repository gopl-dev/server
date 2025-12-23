package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateFindPasswordResetByTokenInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		token     string
	}{
		{
			name:      "empty token",
			expectErr: "Token is required",
			argName:   "token",
			token:     "",
		},
		{
			name:  "valid input",
			valid: true,
			token: "some-valid-token",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateFindPasswordResetByTokenInput(c.token)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
