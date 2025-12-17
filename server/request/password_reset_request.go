package request

import (
	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

// PasswordResetRequest represents the request body for initiating a password reset.
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// ValidationSchema ...
func (r *PasswordResetRequest) ValidationSchema() z.Shape {
	return ds.PasswordResetRequestValidationRules
}

// PasswordReset represents the request body for resetting a password with a token.
type PasswordReset struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ValidationSchema ...
func (r *PasswordReset) ValidationSchema() z.Shape {
	return ds.PasswordResetValidationRules
}
