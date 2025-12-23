package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateChangeUserPasswordInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		data      service.ChangeUserPasswordInput
	}{
		{
			name:      "missing ID",
			expectErr: "userID is required",
			argName:   "user_id",
			data:      service.ChangeUserPasswordInput{0, "aaa", "bbb"},
		},
		{
			name:      "empty old password",
			expectErr: "Password is required",
			argName:   "old_password",
			data:      service.ChangeUserPasswordInput{1, "", "bbb"},
		},
		{
			name:      "empty new password",
			expectErr: "Password is required",
			argName:   "new_password",
			data:      service.ChangeUserPasswordInput{1, "aaa", ""},
		},
		{
			name:      "new password too short",
			expectErr: "Password must be at least 6 characters",
			argName:   "new_password",
			data:      service.ChangeUserPasswordInput{1, "aaa", "bbb"},
		},
		{
			valid: true,
			name:  "valid input",
			data:  service.ChangeUserPasswordInput{1, "aaa", "new-password"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateChangeUserPasswordInput(c.data.UserID, &c.data.OldPassword, &c.data.NewPassword)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
