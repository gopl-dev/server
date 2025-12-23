package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateCreateChangeEmailRequestInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		data      service.CreateChangeEmailRequestInput
	}{
		{
			name:      "missing userID",
			expectErr: "userID is required",
			argName:   "user_id",
			data:      service.CreateChangeEmailRequestInput{0, "mail@ognev.dev"},
		},
		{
			name:      "empty email",
			expectErr: "Email is required",
			argName:   "new_email",
			data:      service.CreateChangeEmailRequestInput{1, ""},
		},
		{
			name:      "invalid email",
			expectErr: "must be a valid email",
			argName:   "new_email",
			data:      service.CreateChangeEmailRequestInput{1, "aaa"},
		},
		{
			valid: true,
			name:  "valid input",
			data:  service.CreateChangeEmailRequestInput{1, "mail@ognev.dev"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateCreateChangeEmailRequestInput(c.data.UserID, &c.data.NewEmail)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
