// Package request ...
package request

import (
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

// UserSignUp ...
type UserSignUp struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Sanitize ...
func (r *UserSignUp) Sanitize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
}

// ValidationSchema ...
func (r *UserSignUp) ValidationSchema() z.Shape {
	return ds.UserValidationRules
}

// ToParams ...
func (r *UserSignUp) ToParams() service.RegisterUserArgs {
	return service.RegisterUserArgs{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
	}
}

// ConfirmEmail ...
type ConfirmEmail struct {
	Code string `json:"code"`
}

// Sanitize ...
func (r *ConfirmEmail) Sanitize() {
	r.Code = strings.TrimSpace(r.Code)
}

// UserSignIn ...
type UserSignIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Sanitize ...
func (r *UserSignIn) Sanitize() {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
}

// ValidationSchema ...
func (r *UserSignIn) ValidationSchema() z.Shape {
	return z.Shape{
		"email":    z.String().Email().Required(z.Message("Email is required")),
		"password": z.String().Required(z.Message("Password is required")),
	}
}
