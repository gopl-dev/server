package request

import (
	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

// ChangePassword represents the request body for changing a user's password.
type ChangePassword struct {
	OldPassword string `json:"old_password" z:"OldPassword"`
	NewPassword string `json:"new_password" z:"NewPassword"`
}

// Sanitize ...
func (r *ChangePassword) Sanitize() {
	// keep it as it
}

// ValidationSchema ...
func (r *ChangePassword) ValidationSchema() z.Shape {
	return ds.ChangePasswordValidationRules
}
