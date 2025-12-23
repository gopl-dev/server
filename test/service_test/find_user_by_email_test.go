package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateFindUserByEmailInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		data      string
	}{
		{
			name:      "invalid email",
			expectErr: "must be a valid email",
			argName:   "email",
			data:      "aaa",
		},
		{
			name:      "empty email",
			expectErr: "Email is required",
			argName:   "email",
			data:      "",
		},
		{
			valid: true,
			name:  "valid input",
			data:  "mail@ognev.dev",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateFindUserByEmailInput(&c.data)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
