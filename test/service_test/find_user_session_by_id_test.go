package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app/service"
)

func TestValidateFindUserSessionByIDInput(t *testing.T) {
	t.Parallel()

	var badUUID uuid.UUID
	copy(badUUID[:], "bad-uuid")

	cases := []struct {
		name      string
		valid     bool
		expectErr string
		argName   string
		id        string
	}{
		{
			name:      "invalid UUID",
			expectErr: "Invalid UUID",
			argName:   "id",
			id:        "badz-uuid",
		},
		{
			valid: true,
			name:  "valid input",
			id:    uuid.New().String(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidateFindUserSessionByIDInput(c.id)
			checkValidatedInput(t, c.valid, err, c.argName, c.expectErr)
		})
	}
}
