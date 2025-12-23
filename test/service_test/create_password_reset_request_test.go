package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateCreatePasswordResetRequestInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		data      service.CreatePasswordResetRequestInput
	}{
		{
			name:      "invalid email",
			expectErr: "must be a valid email",
			argName:   "email",
			data:      service.CreatePasswordResetRequestInput{"aaa"},
		},
		{
			name:      "empty email",
			expectErr: "Email is required",
			argName:   "email",
			data:      service.CreatePasswordResetRequestInput{""},
		},
		{
			valid: true,
			name:  "valid input",
			data:  service.CreatePasswordResetRequestInput{"mail@ognev.dev"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateCreatePasswordResetRequestInput(&c.data.Email)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
