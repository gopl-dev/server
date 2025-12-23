package service_test

import (
	"testing"

	"github.com/gopl-dev/server/app/service"
)

func TestValidateFindUserByIDInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		id        int64
	}{
		{
			name:      "ID missing",
			expectErr: "ID is required",
			argName:   "id",
			id:        0,
		},
		{
			valid: true,
			name:  "valid input",
			id:    1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateFindUserByIDInput(c.id)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
