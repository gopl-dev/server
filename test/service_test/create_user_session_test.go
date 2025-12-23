package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateCreateUserSessionInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		userID    int64
	}{
		{
			name:      "missing ID",
			expectErr: "userID is required",
			argName:   "user_id",
			userID:    0,
		},
		{
			valid:  true,
			name:   "valid input",
			userID: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateCreateUserSessionInput(c.userID)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
