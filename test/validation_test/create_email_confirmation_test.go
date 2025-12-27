package validation_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateCreateEmailConfirmationInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		data      service.CreateEmailConfirmationInput
	}{
		{
			name:      "missing userID",
			expectErr: "userID is required",
			argName:   "user_id",
			data:      service.CreateEmailConfirmationInput{0},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.Normalize(&c.data)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
