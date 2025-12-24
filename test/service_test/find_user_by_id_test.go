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
		data      service.FindUserByIDInput
	}{
		{
			name:      "missing ID",
			expectErr: "ID is required",
			argName:   "id",
			data:      service.FindUserByIDInput{0},
		},
		{
			valid: true,
			name:  "valid input",
			data:  service.FindUserByIDInput{1},
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